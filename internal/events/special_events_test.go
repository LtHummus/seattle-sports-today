package events

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDynamoClient struct {
	responses         []*dynamodb.QueryOutput
	err               error
	calls             int
	receivedInputs    []*dynamodb.QueryInput
	receivedStartKeys []map[string]types.AttributeValue
}

func (f *fakeDynamoClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	f.receivedInputs = append(f.receivedInputs, params)
	f.receivedStartKeys = append(f.receivedStartKeys, params.ExclusiveStartKey)
	if f.err != nil {
		return nil, f.err
	}

	if f.calls >= len(f.responses) {
		return &dynamodb.QueryOutput{}, nil
	}

	resp := f.responses[f.calls]
	f.calls++
	return resp, nil
}

func buildItem(t *testing.T, record specialEventRecord) map[string]types.AttributeValue {
	t.Helper()
	item, err := attributevalue.MarshalMap(record)
	require.NoError(t, err)
	return item
}

func TestSpecialEventsForDate(t *testing.T) {
	prevClient := dynamoClient
	defer func() {
		dynamoClient = prevClient
	}()

	t.Setenv("SPECIAL_EVENTS_TABLE_NAME", "test-table")

	date := time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC)
	formattedDate := date.Format("2006-01-02")

	record1 := specialEventRecord{
		Date:           formattedDate,
		Slug:           "slug-1",
		TeamName:       "Seattle Kraken",
		Venue:          "Climate Pledge Arena",
		LocalTime:      "7:00 PM",
		Opponent:       "Boston Bruins",
		RawDescription: "Kraken vs Bruins",
		RawTime:        111,
	}
	record2 := specialEventRecord{
		Date:           formattedDate,
		Slug:           "slug-2",
		TeamName:       "Seattle Sounders",
		Venue:          "Lumen Field",
		LocalTime:      "1:00 PM",
		Opponent:       "Portland Timbers",
		RawDescription: "Sounders vs Timbers",
		RawTime:        222,
	}

	key1 := map[string]types.AttributeValue{"date": &types.AttributeValueMemberS{Value: formattedDate}, "slug": &types.AttributeValueMemberS{Value: "slug-1"}}
	key2 := map[string]types.AttributeValue{"date": &types.AttributeValueMemberS{Value: formattedDate}, "slug": &types.AttributeValueMemberS{Value: "slug-2"}}

	tests := []struct {
		name              string
		responses         []*dynamodb.QueryOutput
		queryErr          error
		expectedItems     int
		expectedTeams     []string
		expectedCalls     int
		expectedStartKeys []map[string]types.AttributeValue
		expectErr         bool
		errSubstring      string
	}{
		{
			name: "single page results",
			responses: []*dynamodb.QueryOutput{
				{
					Items: []map[string]types.AttributeValue{buildItem(t, record1), buildItem(t, record2)},
				},
			},
			expectedItems:     2,
			expectedTeams:     []string{"Seattle Kraken", "Seattle Sounders"},
			expectedCalls:     1,
			expectedStartKeys: []map[string]types.AttributeValue{nil},
		},
		{
			name: "multiple pages results",
			responses: []*dynamodb.QueryOutput{
				{
					Items:            []map[string]types.AttributeValue{buildItem(t, record1)},
					LastEvaluatedKey: key1,
				},
				{
					Items:            []map[string]types.AttributeValue{buildItem(t, record2)},
					LastEvaluatedKey: key2,
				},
				{
					Items: []map[string]types.AttributeValue{},
				},
			},
			expectedItems:     2,
			expectedTeams:     []string{"Seattle Kraken", "Seattle Sounders"},
			expectedCalls:     3,
			expectedStartKeys: []map[string]types.AttributeValue{nil, key1, key2},
		},
		{
			name: "no results",
			responses: []*dynamodb.QueryOutput{
				{
					Items: []map[string]types.AttributeValue{},
				},
			},
			expectedItems:     0,
			expectedCalls:     1,
			expectedStartKeys: []map[string]types.AttributeValue{nil},
		},
		{
			name:         "query error",
			queryErr:     errors.New("boom"),
			expectErr:    true,
			errSubstring: "could not query dynamo",
		},
		{
			name: "unmarshal error",
			responses: []*dynamodb.QueryOutput{
				{
					Items: []map[string]types.AttributeValue{
						{"raw_time": &types.AttributeValueMemberS{Value: "not-a-number"}}, // raw_time should be N (int64)
					},
				},
			},
			expectErr:    true,
			errSubstring: "could not unmarshal dynamo items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeDynamoClient{
				responses: tt.responses,
				err:       tt.queryErr,
			}
			dynamoClient = fake

			events, err := specialEventsForDate(context.Background(), date)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstring)
				return
			}

			require.NoError(t, err)
			assert.Len(t, events, tt.expectedItems, "Should have %d item(s), but has %d. Events: %v", tt.expectedItems, len(events), events)

			for i, team := range tt.expectedTeams {
				if i < len(events) {
					assert.Equal(t, team, events[i].TeamName)
				}
			}

			assert.Equal(t, tt.expectedCalls, fake.calls, "Number of calls mismatch")
			require.Equal(t, len(tt.expectedStartKeys), len(fake.receivedStartKeys), "Received start keys length mismatch")

			for i, expectedKey := range tt.expectedStartKeys {
				assert.True(t, reflect.DeepEqual(expectedKey, fake.receivedStartKeys[i]), "Start key mismatch at call %d", i)
				assert.Equal(t, "test-table", *fake.receivedInputs[i].TableName)
				assert.Equal(t, "#date = :date", *fake.receivedInputs[i].KeyConditionExpression)
			}
		})
	}
}

func TestGetSpecialEvents(t *testing.T) {
	prevClient := dynamoClient
	defer func() {
		dynamoClient = prevClient
	}()

	t.Setenv("SPECIAL_EVENTS_TABLE_NAME", "test-table")

	today := time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC)
	tomorrow := today.Add(24 * time.Hour)

	fake := &fakeDynamoClient{
		responses: []*dynamodb.QueryOutput{
			{
				Items: []map[string]types.AttributeValue{
					buildItem(t, specialEventRecord{TeamName: "Today Team"}),
				},
			},
			{
				Items: []map[string]types.AttributeValue{
					buildItem(t, specialEventRecord{TeamName: "Tomorrow Team"}),
				},
			},
		},
	}
	dynamoClient = fake

	todayEvents, tomorrowEvents, err := getSpecialEvents(context.Background(), today, tomorrow)
	require.NoError(t, err)

	assert.Len(t, todayEvents, 1)
	assert.Equal(t, "Today Team", todayEvents[0].TeamName)
	assert.Len(t, tomorrowEvents, 1)
	assert.Equal(t, "Tomorrow Team", tomorrowEvents[0].TeamName)

	require.Len(t, fake.receivedInputs, 2)
	assert.Equal(t, today.Format("2006-01-02"), fake.receivedInputs[0].ExpressionAttributeValues[":date"].(*types.AttributeValueMemberS).Value)
	assert.Equal(t, tomorrow.Format("2006-01-02"), fake.receivedInputs[1].ExpressionAttributeValues[":date"].(*types.AttributeValueMemberS).Value)
}

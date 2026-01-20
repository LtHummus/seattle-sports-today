package events

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

const tableEnvironmentVariableName = "SPECIAL_EVENTS_TABLE_NAME"

type dynamoQueryAPI interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

var dynamoClient dynamoQueryAPI

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not load AWS config")
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	log.Info().Str("table_name", os.Getenv(tableEnvironmentVariableName)).Msg("initialized dynamodb client")
}

type specialEventRecord struct {
	Date           string `dynamodbav:"date"`
	Slug           string `dynamodbav:"slug"`
	TeamName       string `dynamodbav:"team_name"`
	Venue          string `dynamodbav:"venue"`
	LocalTime      string `dynamodbav:"local_time"`
	Opponent       string `dynamodbav:"opponent"`
	RawDescription string `dynamodbav:"raw_description"`
	RawTime        int64  `dynamodbav:"raw_time"`
}

func specialEventsForDate(ctx context.Context, t time.Time) ([]*Event, error) {
	formattedDate := t.Format("2006-01-02")

	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv(tableEnvironmentVariableName)),
		KeyConditionExpression: aws.String("#date = :date"),
		ExpressionAttributeNames: map[string]string{
			"#date": "date",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":date": &types.AttributeValueMemberS{Value: formattedDate},
		},
	}

	var events []*Event

	paginator := dynamodb.NewQueryPaginator(dynamoClient, queryInput)

	// Strategy: Return partially processed results to caller even on error, they might want to do something with it
	for paginator.HasMorePages() {
		res, err := paginator.NextPage(ctx)
		if err != nil {
			return events, fmt.Errorf("events: getSpecialEvents: could not query dynamo: %w", err)
		}

		var pageItems []specialEventRecord
		err = attributevalue.UnmarshalListOfMaps(res.Items, &pageItems)
		if err != nil {
			return events, fmt.Errorf("events: getSpecialEvents: could not unmarshal dynamo items: %w", err)
		}

		for _, curr := range pageItems {
			events = append(events, &Event{
				TeamName:       curr.TeamName,
				Venue:          curr.Venue,
				LocalTime:      curr.LocalTime,
				Opponent:       curr.Opponent,
				RawDescription: curr.RawDescription,
				RawTime:        curr.RawTime,
			})
		}
	}

	return events, nil
}

func getSpecialEvents(ctx context.Context, today time.Time, tomorrow time.Time) ([]*Event, []*Event, error) {
	todayEvents, err := specialEventsForDate(ctx, today)
	if err != nil {
		return nil, nil, err
	}
	tomorrowEvents, err := specialEventsForDate(ctx, tomorrow)
	if err != nil {
		return nil, nil, err
	}

	return todayEvents, tomorrowEvents, nil
}

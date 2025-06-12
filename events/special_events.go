package events

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/rs/zerolog/log"
)

const tableEnvironmentVariableName = "SPECIAL_EVENTS_TABLE_NAME"

var dynamoClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not load AWS config")
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	log.Info().Str("table_name", os.Getenv(tableEnvironmentVariableName)).Msg("initialized dynamodb client")
}

func GetSpecialEvents(ctx context.Context) ([]*Event, error) {
	formattedDate := SeattleCurrentTime.Format("2006-01-02")

	res, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv(tableEnvironmentVariableName)),
		KeyConditionExpression: aws.String("#date = :date"),
		ExpressionAttributeNames: map[string]string{
			"#date": "date",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":date": &types.AttributeValueMemberS{Value: formattedDate},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("events: GetSpecialEvents: could not query dynamo: %w", err)
	}

	var items []struct {
		Date           string `dynamodbav:"date"`
		Slug           string `dynamodbav:"slug"`
		TeamName       string `dynamodbav:"team_name"`
		Venue          string `dynamodbav:"venue"`
		LocalTime      string `dynamodbav:"local_time"`
		Opponent       string `dynamodbav:"opponent"`
		RawDescription string `dynamodbav:"raw_description"`
		RawTime        int64  `dynamodbav:"raw_time"`
	}
	err = attributevalue.UnmarshalListOfMaps(res.Items, &items)
	if err != nil {
		return nil, fmt.Errorf("events: GetSpecialEvents: could not unmarshal dynamo items: %w", err)
	}

	ret := make([]*Event, len(items))
	for i, curr := range items {
		ret[i] = &Event{
			TeamName:       curr.TeamName,
			Venue:          curr.Venue,
			LocalTime:      curr.LocalTime,
			Opponent:       curr.Opponent,
			RawDescription: curr.RawDescription,
			RawTime:        curr.RawTime,
		}
	}

	return ret, nil
}

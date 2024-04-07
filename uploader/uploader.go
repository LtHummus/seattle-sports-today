package uploader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

const (
	envVarBucketName     = "UPLOAD_S3_BUCKET_NAME"
	envVarDistributionID = "UPLOAD_CF_DISTRIBUTION_ID"
)

func Upload(ctx context.Context, contents []byte) error {
	log.Info().Msg("loading AWS config")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("uploader: Upload: could not load AWS config: %w", err)
	}

	log.Info().Msg("creating clients")
	s3Client := s3.NewFromConfig(cfg)
	cfClient := cloudfront.NewFromConfig(cfg)

	bucketName := os.Getenv(envVarBucketName)
	log.Info().Str("bucket", bucketName).Msg("uploading object")

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String("index.html"),
		Body:        bytes.NewReader(contents),
		ContentType: aws.String("text/html"),
	})
	if err != nil {
		return fmt.Errorf("uploader: Upload: could not upload to S3: %w", err)
	}

	distributionID := os.Getenv(envVarDistributionID)
	log.Info().Str("distribution_id", distributionID).Msg("invalidating CF cache")
	_, err = cfClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339)),
			Paths: &types.Paths{
				Items:    []string{"/*"},
				Quantity: aws.Int32(1),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("uploader: Upload: could not invalidate cache: %w", err)
	}

	log.Info().Msg("upload done")

	return nil

}

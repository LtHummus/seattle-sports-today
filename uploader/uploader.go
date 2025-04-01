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
	"golang.org/x/sync/errgroup"
)

const (
	envVarBucketName     = "UPLOAD_S3_BUCKET_NAME"
	envVarDistributionID = "UPLOAD_CF_DISTRIBUTION_ID"
)

var (
	s3Client *s3.Client
	cfClient *cloudfront.Client
)

func init() {
	log.Info().Msg("loading AWS config for uploader")
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not load AWS config")
	}

	log.Info().Msg("creating clients")
	s3Client = s3.NewFromConfig(cfg)
	cfClient = cloudfront.NewFromConfig(cfg)
}

func uploadObject(ctx context.Context, contents []byte, key string, contentType string) error {
	bucketName := os.Getenv(envVarBucketName)
	log.Info().Str("bucket", bucketName).Str("key", key).Str("content_type", contentType).Msg("uploading object")

	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(contents),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("uploader: Upload: could not upload to S3: %s: %w", key, err)
	}

	return nil
}

func Upload(ctx context.Context, htmlContents []byte, jsonContents []byte) error {
	eg, ctx2 := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return uploadObject(ctx2, htmlContents, "index.html", "text/html")
	})
	eg.Go(func() error {
		return uploadObject(ctx2, jsonContents, "todays_events.json", "application/json")
	})

	err := eg.Wait()
	if err != nil {
		return fmt.Errorf("upload: could not upload objects to S3: %w", err)
	}

	distributionID := os.Getenv(envVarDistributionID)
	log.Info().Str("distribution_id", distributionID).Msg("invalidating CF cache")
	_, err = cfClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
		DistributionId: aws.String(distributionID),
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: aws.String(time.Now().Format(time.RFC3339)),
			Paths: &types.Paths{
				Items:    []string{"/index.html", "/todays_events.json"},
				Quantity: aws.Int32(2),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("uploader: Upload: could not invalidate cache: %w", err)
	}

	log.Info().Msg("upload done")

	return nil

}

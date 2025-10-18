package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/rs/zerolog/log"
)

var secretsManagerClient *secretsmanager.Client

func Init(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	secretsManagerClient = secretsmanager.NewFromConfig(cfg)

	return nil
}

func GetSecretString(ctx context.Context, secretName string) (string, error) {
	log.Info().Str("secret_name", secretName).Msg("loading secret")
	res, err := secretsManagerClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return "", err
	}

	return *res.SecretString, nil
}

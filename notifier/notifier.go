package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/rs/zerolog/log"
)

const envVarNotificationSecretName = "NOTIFIER_SECRET_NAME"

type Priority int

const (
	_ Priority = iota
	PriorityMin
	PriorityLow
	PriorityDefault
	PriorityHigh
	PriorityMax
)

type Emoji string

const (
	EmojiNone  Emoji = ""
	EmojiSiren Emoji = "rotating_light"
	EmojiParty Emoji = "partying_face"
)

var httpClient = xray.Client(http.DefaultClient)

func Notify(ctx context.Context, text string, priority Priority, emoji Emoji) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not load AWS config")
		return fmt.Errorf("notifier: Notify: could not load AWS config: %w", err)
	}

	secretClient := secretsmanager.NewFromConfig(cfg)
	secretARN := os.Getenv(envVarNotificationSecretName)

	log.Info().Str("secret_arn", secretARN).Msg("loading secret")

	res, err := secretClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		log.Error().Str("secret_arn", secretARN).Err(err).Msg("could not retrieve secret")
		return fmt.Errorf("notifier: Notify: could not retrieve secret: %w", err)
	}

	notifierKey := *res.SecretString

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://ntfy.sh/%s", notifierKey), strings.NewReader(text))
	if err != nil {
		log.Error().Err(err).Msg("could not build ntfy request")
		return fmt.Errorf("notifier: Notify: could not build request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Title", "Seattle Sports Today Notification")
	req.Header.Set("Priority", fmt.Sprintf("%d", priority))
	if emoji != EmojiNone {
		req.Header.Set("Tags", string(emoji))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("could not make ntfy request")
		return fmt.Errorf("notifier: Notify: could not make request: %w", err)
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)
	return nil
}

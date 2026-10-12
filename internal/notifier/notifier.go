package notifier

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-xray-sdk-go/v2/xray"
	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/secrets"
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

var httpClient = xray.Client(&http.Client{
	Timeout: 10 * time.Second,
})

func Notify(ctx context.Context, text string, priority Priority, emoji Emoji) error {
	secretARN := os.Getenv(envVarNotificationSecretName)

	notifierKey, err := secrets.GetSecretString(ctx, secretARN)

	if err != nil {
		log.Error().Err(err).Msg("could not retrieve notifier secret")
		return fmt.Errorf("notifier: Notify: could not retrieve secret: %w", err)
	}

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
		// *url.Error includes the request URL, which contains our secret topic name
		if uerr, ok := errors.AsType[*url.Error](err); ok {
			err = uerr.Err
		}
		log.Error().Err(err).Msg("could not make ntfy request")
		return fmt.Errorf("notifier: Notify: could not make request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Warn().Err(err).Msg("error closing http response from ntfy")
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Error().Str("status", resp.Status).Str("body", string(body)).Msg("ntfy rejected notification")
		return fmt.Errorf("notifier: Notify: ntfy returned %s: %s", resp.Status, string(body))
	}

	return nil
}

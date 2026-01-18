package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/events"
	"github.com/lthummus/seattle-sports-today/internal/notifier"
	"github.com/lthummus/seattle-sports-today/internal/renderhtml"
	"github.com/lthummus/seattle-sports-today/internal/renderjson"
	"github.com/lthummus/seattle-sports-today/internal/secrets"
	"github.com/lthummus/seattle-sports-today/internal/uploader"
)

type CustomEvent struct {
	Today  string `json:"today"`
	Upload bool   `json:"upload"`
}

func eventHandler(ctx context.Context, event CustomEvent) error {
	defer func() {
		if err := recover(); err != nil {
			_ = notifier.Notify(context.Background(), fmt.Sprintf("ERROR: uncaught panic: %v", err), notifier.PriorityHigh, notifier.EmojiSiren)
		}
	}()

	var triggeredByEventBridge bool
	var seattleToday time.Time
	var seattleTomorrow time.Time

	if event.Today == "" {
		// no event data, assume it was triggered by event bridge
		log.Info().Msg("no custom event data; assuming event bridge")
		triggeredByEventBridge = true

		seattleToday = time.Now().In(events.SeattleTimeZone)
	} else {
		var err error
		seattleToday, err = time.ParseInLocation("2006-01-02", event.Today, events.SeattleTimeZone)
		if err != nil {
			return fmt.Errorf("invalid input date format: %s: %w", event.Today, err)
		}
	}

	seattleTomorrow = seattleToday.AddDate(0, 0, 1)

	log.Info().Bool("triggered_by_event_bridge", triggeredByEventBridge).Time("seattle_today", seattleToday).Time("seattle_tomorrow", seattleTomorrow).Msg("getting today's games")
	eventResults, err := events.GetTodayAndTomorrowGames(ctx, seattleToday, seattleTomorrow)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not get today's games: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	log.Info().Int("today_games_found", len(eventResults.TodayEvent)).Int("tomorrow_games_found", len(eventResults.TomorrowEvents)).Msg("found games")

	slices.SortFunc(eventResults.TodayEvent, func(a, b *events.Event) int {
		return int(a.RawTime - b.RawTime)
	})
	slices.SortFunc(eventResults.TomorrowEvents, func(a, b *events.Event) int {
		return int(a.RawTime - b.RawTime)
	})

	for _, curr := range eventResults.TodayEvent {
		log.Info().Str("team_name", curr.TeamName).Str("venue", curr.Venue).Str("local_time", curr.LocalTime).Str("opponent", curr.Opponent).Int64("raw_time", curr.RawTime).Msg("found event today")
	}
	for _, curr := range eventResults.TomorrowEvents {
		log.Info().Str("team_name", curr.TeamName).Str("venue", curr.Venue).Str("local_time", curr.LocalTime).Str("opponent", curr.Opponent).Int64("raw_time", curr.RawTime).Msg("found event tomorrow")
	}

	log.Info().Msg("rendering page")
	page, err := renderhtml.RenderPage(eventResults, seattleToday)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not render page: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	jsonData, err := renderjson.RenderJSON(eventResults, seattleToday)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not render JSON: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	log.Info().Msg("render complete")

	runningInDefaultMode := os.Getenv("_HANDLER") != "" && triggeredByEventBridge
	shouldUploadAnyway := os.Getenv("UPLOAD_ANYWAY") == "true" || (!triggeredByEventBridge && event.Upload)

	if runningInDefaultMode || shouldUploadAnyway {
		log.Info().Msg("beginning upload")
		err = uploader.Upload(ctx, page, jsonData)
		if err != nil {
			_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: upload page: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
			return err
		}

		log.Info().Msg("upload complete")
	} else {
		log.Warn().Msg("detected running locally, not uploading")
		fmt.Printf("%s\n----------\n%s\n", string(jsonData), string(page))
	}

	log.Info().Msg("all in a day's work...")

	notificationMessage := fmt.Sprintf("Everything worked! Found %d game(s) for %s and %d game(s) for %s",
		len(eventResults.TodayEvent),
		seattleToday.Format("2006-04-02"),
		len(eventResults.TomorrowEvents),
		seattleTomorrow.Format("2006-04-02"))

	err = notifier.Notify(ctx, notificationMessage, notifier.PriorityDefault, notifier.EmojiParty)
	if err != nil {
		log.Warn().Err(err).Msg("error sending notification")
	}

	return nil
}

func main() {
	log.Info().Msg("hello world")

	err := secrets.Init(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize secrets manager client")
	}

	if os.Getenv("_HANDLER") != "" {
		lambda.Start(eventHandler)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		ce := CustomEvent{}
		if todayDateString := os.Getenv("TEST_DATE"); todayDateString != "" {
			ce.Today = todayDateString
		}
		err := eventHandler(context.Background(), ce)
		if err != nil {
			log.Error().Err(err).Msg("error running handler")
		}
	}
}

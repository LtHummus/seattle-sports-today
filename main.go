package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"slices"

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

func eventHandler(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			_ = notifier.Notify(context.Background(), fmt.Sprintf("ERROR: uncaught panic: %v", err), notifier.PriorityHigh, notifier.EmojiSiren)
		}
	}()

	log.Info().Msg("getting today's games")
	eventResults, err := events.GetTodaysGames(ctx)
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
	page, err := renderhtml.RenderPage(eventResults)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not render page: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	jsonData, err := renderjson.RenderJSON(eventResults)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not render JSON: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	log.Info().Msg("render complete")

	if os.Getenv("_HANDLER") != "" || os.Getenv("UPLOAD_ANYWAY") == "true" {
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

	err = notifier.Notify(ctx, fmt.Sprintf("Everything worked! Found %d game(s) today and %d game(s) tomorrow", len(eventResults.TodayEvent), len(eventResults.TomorrowEvents)), notifier.PriorityDefault, notifier.EmojiParty)
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
		err := eventHandler(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("error running handler")
		}
	}
}

package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/events"
	"github.com/lthummus/seattle-sports-today/notifier"
	"github.com/lthummus/seattle-sports-today/uploader"
)

//go:embed index.gohtml
var templateString string

var pageTemplate *template.Template

type templateParams struct {
	Events        []*events.Event
	GeneratedDate string
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: true})

	var err error
	pageTemplate, err = template.New("").Parse(templateString)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse template")
	}
}

func renderPage(foundEvents []*events.Event) ([]byte, error) {
	generatedDateString := events.SeattleCurrentTime.Format("Monday Jan _2, 2006")
	log.Info().Int("num_events", len(foundEvents)).Str("formatted_date", generatedDateString).Msg("rendering page")
	buf := bytes.NewBuffer(nil)
	err := pageTemplate.Execute(buf, &templateParams{Events: foundEvents, GeneratedDate: generatedDateString})
	if err != nil {
		return nil, fmt.Errorf("renderPage: could not render: %w", err)
	}

	return buf.Bytes(), nil
}

func eventHandler(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			_ = notifier.Notify(context.Background(), fmt.Sprintf("ERROR: uncaught panic: %v", err), notifier.PriorityHigh, notifier.EmojiSiren)
		}
	}()

	log.Info().Msg("getting today's games")
	events, err := events.GetTodaysGames(ctx)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not get today's games: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}
	
	log.Info().Int("games_found", len(events)).Msg("found games")

	log.Info().Msg("rendering page")
	page, err := renderPage(events)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: could not render page: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	log.Info().Msg("render complete")

	log.Info().Msg("beginning upload")
	err = uploader.Upload(ctx, page)
	if err != nil {
		_ = notifier.Notify(ctx, fmt.Sprintf("ERROR: upload page: %s", err.Error()), notifier.PriorityHigh, notifier.EmojiSiren)
		return err
	}

	log.Info().Msg("upload complete")

	log.Info().Msg("all in a day's work...")

	err = notifier.Notify(ctx, fmt.Sprintf("Everything worked! Found %d game(s)", len(events)), notifier.PriorityDefault, notifier.EmojiParty)
	if err != nil {
		log.Warn().Err(err).Msg("error sending notification")
	}

	return nil
}

func main() {
	log.Info().Msg("hello world")

	if os.Getenv("_HANDLER") != "" {
		lambda.Start(eventHandler)
	} else {
		eventHandler(context.Background())
	}
}

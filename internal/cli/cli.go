package cli

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	urfavecli "github.com/urfave/cli/v3"

	"github.com/lthummus/seattle-sports-today/internal/handler"
)

var (
	testDate        string
	uploadAnyway    bool
	invalidateCache bool

	rootCmd *urfavecli.Command
)

func init() {
	seattleTimeZone, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal().Err(err).Msg("could not load seattle time zone")
	}

	rootCmd = &urfavecli.Command{
		Name:  "seattle-sports-today",
		Usage: "render and generate the page for isthereaseattlehomegametoday.com",
		Flags: []urfavecli.Flag{
			&urfavecli.BoolFlag{
				Name:        "upload-anyway",
				Value:       false,
				Usage:       "upload anyway even if we're running locally",
				Destination: &uploadAnyway,
			},
			&urfavecli.BoolFlag{
				Name:        "invalidate-all",
				Value:       false,
				Usage:       "invalidate everything in cloudfront cache (requires upload anyway)",
				Destination: &invalidateCache,
			},
			&urfavecli.StringFlag{
				Name:        "date",
				Value:       time.Now().In(seattleTimeZone).Format("2006-01-02"),
				Usage:       "Date to run for (defaults to today)",
				Destination: &testDate,
			},
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
			ce := handler.CustomEvent{}
			ce.Today = testDate
			if uploadAnyway {
				ce.Upload = true
			}
			if invalidateCache {
				ce.InvalidateAll = true
			}
			err := handler.EventHandler(ctx, ce)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func Execute(ctx context.Context, args []string) {
	if err := rootCmd.Run(ctx, args); err != nil {
		log.Fatal().Err(err).Msg("an error happened running the handler")
	}
}

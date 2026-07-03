package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/internal/cli"
	"github.com/lthummus/seattle-sports-today/internal/handler"
	"github.com/lthummus/seattle-sports-today/internal/secrets"
)

func main() {
	log.Info().Msg("hello world")

	err := secrets.Init(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize secrets manager client")
	}

	if os.Getenv("_HANDLER") != "" {
		lambda.Start(handler.EventHandler)
	} else {
		cli.Execute(context.Background(), os.Args)
	}
}

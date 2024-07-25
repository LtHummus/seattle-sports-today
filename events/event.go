package events

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/rs/zerolog/log"
)

var (
	SeattleTimeZone    *time.Location
	SeattleCurrentTime time.Time
)

const (
	localTimeDateFormat = "3:04 PM"
)

func init() {
	stz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal().Err(err).Msg("could not load seattle timezeone")
	}

	SeattleTimeZone = stz
	SeattleCurrentTime = time.Now().In(SeattleTimeZone)

	log.Info().Str("current_seattle_time", SeattleCurrentTime.Format(time.RFC850)).Str("seattle_time_zone", SeattleTimeZone.String()).Msg("initialized time")
}

type Event struct {
	TeamName  string `json:"string"`
	Venue     string `json:"venue"`
	LocalTime string `json:"local_time"`
	Opponent  string `json:"opponent"`
}

func GetTodaysGames(ctx context.Context) ([]*Event, error) {

	eg, ctx2 := errgroup.WithContext(ctx)
	var soundersGame *Event
	var marinersGame *Event
	var krakenGame *Event
	var seahawksGame *Event
	var stormGame *Event
	var reignGame *Event
	var huskiesFootball *Event

	eg.Go(func() error {
		var e error
		soundersGame, e = GetSoundersGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		marinersGame, e = GetMarinersGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		krakenGame, e = GetKrakenGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		seahawksGame, e = GetSeahawksGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		stormGame, e = GetStormGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		reignGame, e = GetReignGame(ctx2)
		return e
	})

	eg.Go(func() error {
		var e error
		huskiesFootball, e = GetHuskiesFootballGame(ctx2)
		return e
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	var events []*Event
	if soundersGame != nil {
		events = append(events, soundersGame)
	}
	if marinersGame != nil {
		events = append(events, marinersGame)
	}
	if krakenGame != nil {
		events = append(events, krakenGame)
	}
	if seahawksGame != nil {
		events = append(events, seahawksGame)
	}
	if stormGame != nil {
		events = append(events, stormGame)
	}
	if reignGame != nil {
		events = append(events, reignGame)
	}
	if huskiesFootball != nil {
		events = append(events, huskiesFootball)
	}

	return events, nil
}

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const marinersTeamID = 136

type mlbTodayGamesResponse struct {
	Copyright            string `json:"copyright"`
	TotalItems           int    `json:"totalItems"`
	TotalEvents          int    `json:"totalEvents"`
	TotalGames           int    `json:"totalGames"`
	TotalGamesInProgress int    `json:"totalGamesInProgress"`
	Dates                []struct {
		Date                 string `json:"date"`
		TotalItems           int    `json:"totalItems"`
		TotalEvents          int    `json:"totalEvents"`
		TotalGames           int    `json:"totalGames"`
		TotalGamesInProgress int    `json:"totalGamesInProgress"`
		Games                []struct {
			GamePk       int       `json:"gamePk"`
			GameGuid     string    `json:"gameGuid"`
			Link         string    `json:"link"`
			GameType     string    `json:"gameType"`
			Season       string    `json:"season"`
			GameDate     time.Time `json:"gameDate"`
			OfficialDate string    `json:"officialDate"`
			Status       struct {
				AbstractGameState string `json:"abstractGameState"`
				CodedGameState    string `json:"codedGameState"`
				DetailedState     string `json:"detailedState"`
				StatusCode        string `json:"statusCode"`
				StartTimeTBD      bool   `json:"startTimeTBD"`
				AbstractGameCode  string `json:"abstractGameCode"`
			} `json:"status"`
			Teams struct {
				Away struct {
					LeagueRecord struct {
						Wins   int    `json:"wins"`
						Losses int    `json:"losses"`
						Pct    string `json:"pct"`
					} `json:"leagueRecord"`
					Score int `json:"score"`
					Team  struct {
						Id   int    `json:"id"`
						Name string `json:"name"`
						Link string `json:"link"`
					} `json:"team"`
					IsWinner     bool `json:"isWinner,omitempty"`
					SplitSquad   bool `json:"splitSquad"`
					SeriesNumber int  `json:"seriesNumber"`
				} `json:"away"`
				Home struct {
					LeagueRecord struct {
						Wins   int    `json:"wins"`
						Losses int    `json:"losses"`
						Pct    string `json:"pct"`
					} `json:"leagueRecord"`
					Score int `json:"score"`
					Team  struct {
						Id   int    `json:"id"`
						Name string `json:"name"`
						Link string `json:"link"`
					} `json:"team"`
					IsWinner     bool `json:"isWinner,omitempty"`
					SplitSquad   bool `json:"splitSquad"`
					SeriesNumber int  `json:"seriesNumber"`
				} `json:"home"`
			} `json:"teams"`
			Venue struct {
				Id   int    `json:"id"`
				Name string `json:"name"`
				Link string `json:"link"`
			} `json:"venue"`
			Content struct {
				Link string `json:"link"`
			} `json:"content"`
			IsTie                  bool   `json:"isTie,omitempty"`
			GameNumber             int    `json:"gameNumber"`
			PublicFacing           bool   `json:"publicFacing"`
			DoubleHeader           string `json:"doubleHeader"`
			GamedayType            string `json:"gamedayType"`
			Tiebreaker             string `json:"tiebreaker"`
			CalendarEventID        string `json:"calendarEventID"`
			SeasonDisplay          string `json:"seasonDisplay"`
			DayNight               string `json:"dayNight"`
			ScheduledInnings       int    `json:"scheduledInnings"`
			ReverseHomeAwayStatus  bool   `json:"reverseHomeAwayStatus"`
			InningBreakLength      int    `json:"inningBreakLength"`
			GamesInSeries          int    `json:"gamesInSeries"`
			SeriesGameNumber       int    `json:"seriesGameNumber"`
			SeriesDescription      string `json:"seriesDescription"`
			RecordSource           string `json:"recordSource"`
			IfNecessary            string `json:"ifNecessary"`
			IfNecessaryDescription string `json:"ifNecessaryDescription"`
		} `json:"games"`
		Events []interface{} `json:"events"`
	} `json:"dates"`
}

func GetMarinersGame(ctx context.Context) (*Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://statsapi.mlb.com/api/v1/schedule/games/?sportId=1", nil)
	if err != nil {
		return nil, fmt.Errorf("events: GetMarinersGame: could not build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events: GetMarinersGame: could not get data from MLB API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetMarinersGame: could not read error body: %w", err)
		}
		return nil, fmt.Errorf("events: GetMarinersGame: could not retireve MLB schedule: %s", string(body))
	}

	var payload mlbTodayGamesResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("events: GetMarinersGame: could not decode MLB response: %w", err)
	}

	for _, curr := range payload.Dates[0].Games {
		if curr.Teams.Home.Team.Id == marinersTeamID {
			startTime := curr.GameDate.In(seattleTimeZone).Format(localTimeDateFormat)

			return &Event{
				TeamName:  "Seattle Mariners",
				Venue:     curr.Venue.Name,
				LocalTime: startTime,
				Opponent:  curr.Teams.Away.Team.Name,
			}, nil
		}
	}

	return nil, nil

}

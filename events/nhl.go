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

type nhlSeasonResponse struct {
	PreviousSeason int    `json:"previousSeason"`
	CurrentSeason  int    `json:"currentSeason"`
	ClubTimezone   string `json:"clubTimezone"`
	ClubUTCOffset  string `json:"clubUTCOffset"`
	Games          []struct {
		Id       int    `json:"id"`
		Season   int    `json:"season"`
		GameType int    `json:"gameType"`
		GameDate string `json:"gameDate"`
		Venue    struct {
			Default string `json:"default"`
			Fr      string `json:"fr,omitempty"`
			Es      string `json:"es,omitempty"`
		} `json:"venue"`
		NeutralSite       bool      `json:"neutralSite"`
		StartTimeUTC      time.Time `json:"startTimeUTC"`
		EasternUTCOffset  string    `json:"easternUTCOffset"`
		VenueUTCOffset    string    `json:"venueUTCOffset"`
		VenueTimezone     string    `json:"venueTimezone"`
		GameState         string    `json:"gameState"`
		GameScheduleState string    `json:"gameScheduleState"`
		TvBroadcasts      []struct {
			Id             int    `json:"id"`
			Market         string `json:"market"`
			CountryCode    string `json:"countryCode"`
			Network        string `json:"network"`
			SequenceNumber int    `json:"sequenceNumber"`
		} `json:"tvBroadcasts"`
		AwayTeam struct {
			Id        int `json:"id"`
			PlaceName struct {
				Default string `json:"default"`
				Fr      string `json:"fr,omitempty"`
			} `json:"placeName"`
			Abbrev         string `json:"abbrev"`
			Logo           string `json:"logo"`
			DarkLogo       string `json:"darkLogo"`
			AwaySplitSquad bool   `json:"awaySplitSquad"`
			AirlineLink    string `json:"airlineLink,omitempty"`
			AirlineDesc    string `json:"airlineDesc,omitempty"`
			Score          int    `json:"score,omitempty"`
			HotelLink      string `json:"hotelLink,omitempty"`
			HotelDesc      string `json:"hotelDesc,omitempty"`
			RadioLink      string `json:"radioLink,omitempty"`
		} `json:"awayTeam"`
		HomeTeam struct {
			Id        int `json:"id"`
			PlaceName struct {
				Default string `json:"default"`
				Fr      string `json:"fr,omitempty"`
			} `json:"placeName"`
			Abbrev         string `json:"abbrev"`
			Logo           string `json:"logo"`
			DarkLogo       string `json:"darkLogo"`
			HomeSplitSquad bool   `json:"homeSplitSquad"`
			Score          int    `json:"score,omitempty"`
			AirlineLink    string `json:"airlineLink,omitempty"`
			AirlineDesc    string `json:"airlineDesc,omitempty"`
			HotelLink      string `json:"hotelLink,omitempty"`
			HotelDesc      string `json:"hotelDesc,omitempty"`
			RadioLink      string `json:"radioLink,omitempty"`
		} `json:"homeTeam"`
		PeriodDescriptor struct {
			PeriodType string `json:"periodType,omitempty"`
		} `json:"periodDescriptor"`
		GameOutcome struct {
			LastPeriodType string `json:"lastPeriodType"`
		} `json:"gameOutcome,omitempty"`
		WinningGoalie struct {
			PlayerId     int `json:"playerId"`
			FirstInitial struct {
				Default string `json:"default"`
			} `json:"firstInitial"`
			LastName struct {
				Default string `json:"default"`
				Cs      string `json:"cs,omitempty"`
				Sk      string `json:"sk,omitempty"`
				Fi      string `json:"fi,omitempty"`
			} `json:"lastName"`
		} `json:"winningGoalie,omitempty"`
		WinningGoalScorer struct {
			PlayerId     int `json:"playerId"`
			FirstInitial struct {
				Default string `json:"default"`
			} `json:"firstInitial"`
			LastName struct {
				Default string `json:"default"`
				Cs      string `json:"cs,omitempty"`
				Sk      string `json:"sk,omitempty"`
				Fi      string `json:"fi,omitempty"`
			} `json:"lastName"`
		} `json:"winningGoalScorer,omitempty"`
		GameCenterLink  string `json:"gameCenterLink"`
		ThreeMinRecap   string `json:"threeMinRecap,omitempty"`
		ThreeMinRecapFr string `json:"threeMinRecapFr,omitempty"`
		SpecialEvent    struct {
			Default string `json:"default"`
			Fr      string `json:"fr"`
		} `json:"specialEvent,omitempty"`
		SpecialEventLogo string `json:"specialEventLogo,omitempty"`
		TicketsLink      string `json:"ticketsLink,omitempty"`
	} `json:"games"`
}

var nhlTeamMap = map[string]string{
	"ATL": "Atlanta Thrashers",
	"HFD": "Hartford Whalers",
	"MNS": "Minnesota North Stars",
	"QUE": "Quebec Nordiques",
	"WIN": "Winnipeg Jets (1979)",
	"CLR": "Colorado Rockies",
	"SEN": "Ottawa Senators (1917)",
	"HAM": "Hamilton Tigers",
	"PIR": "Pittsburgh Pirates",
	"QUA": "Philadelphia Quakers",
	"DCG": "Detroit Cougars",
	"MWN": "Montreal Wanderers",
	"QBD": "Quebec Bulldogs",
	"MMR": "Montreal Maroons",
	"NYA": "New York Americans",
	"SLE": "St. Louis Eagles",
	"OAK": "Oakland Seals",
	"AFM": "Atlanta Flames",
	"KCS": "Kansas City Scouts",
	"CLE": "Cleveland Barons",
	"DFL": "Detroit Falcons",
	"BRK": "Brooklyn Americans",
	"CGS": "California Golden Seals",
	"TAN": "Toronto Arenas",
	"TSP": "Toronto St. Patricks",
	"NHL": "NHL",
	"DET": "Detroit Red Wings",
	"BOS": "Boston Bruins",
	"WPG": "Winnipeg Jets",
	"SJS": "San Jose Sharks",
	"PIT": "Pittsburgh Penguins",
	"TBL": "Tampa Bay Lightning",
	"PHI": "Philadelphia Flyers",
	"TOR": "Toronto Maple Leafs",
	"BUF": "Buffalo Sabres",
	"CAR": "Carolina Hurricanes",
	"ARI": "Arizona Coyotes",
	"CGY": "Calgary Flames",
	"MTL": "Montréal Canadiens",
	"WSH": "Washington Capitals",
	"VAN": "Vancouver Canucks",
	"COL": "Colorado Avalanche",
	"NSH": "Nashville Predators",
	"ANA": "Anaheim Ducks",
	"VGK": "Vegas Golden Knights",
	"SEA": "Seattle Kraken",
	"DAL": "Dallas Stars",
	"PHX": "Phoenix Coyotes",
	"CHI": "Chicago Blackhawks",
	"NYR": "New York Rangers",
	"CBJ": "Columbus Blue Jackets",
	"FLA": "Florida Panthers",
	"EDM": "Edmonton Oilers",
	"MIN": "Minnesota Wild",
	"STL": "St. Louis Blues",
	"OTT": "Ottawa Senators",
	"NYI": "New York Islanders",
	"LAK": "Los Angeles Kings",
	"NJD": "New Jersey Devils",
	"TBD": "To be determined",
}

func getCurrentNHLSeason() string {
	today := time.Now()
	thisYear := today.Year()
	if today.Month() > time.August {
		return fmt.Sprintf("%d%d", thisYear, thisYear+1)
	} else {
		return fmt.Sprintf("%d%d", thisYear-1, thisYear)
	}
}

func GetKrakenGame(ctx context.Context) (*Event, error) {
	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://api-web.nhle.com/v1/club-schedule-season/SEA/%s", getCurrentNHLSeason())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events: GetKrakenGame: could not retireve NHL API data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status_code", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: GetKrakenGame: could not read error body: %w", err)
		}
		return nil, fmt.Errorf("events: GetKrakenGame: could not retireve NHL schedule: %s", string(body))
	}

	var payload nhlSeasonResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, fmt.Errorf("events: GetKrakenGame: could not decode NHL response: %w", err)
	}

	for _, curr := range payload.Games {
		if curr.HomeTeam.Abbrev != "SEA" || today != curr.GameDate {
			continue
		}

		startTimeLocal := curr.StartTimeUTC.In(seattleTimeZone).Format(localTimeDateFormat)

		return &Event{
			TeamName:  "Seattle Kraken",
			Venue:     curr.Venue.Default,
			LocalTime: startTimeLocal,
			Opponent:  nhlTeamMap[curr.AwayTeam.Abbrev],
		}, nil
	}

	return nil, nil
}

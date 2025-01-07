package events

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/seattle-sports-today/secrets"
)

type TicketmasterEventSearchResponse struct {
	Embedded struct {
		Events []TicketmasterEvent `json:"events"`
	} `json:"_embedded"`
}

type TicketmasterEvent struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Id     string `json:"id"`
	Test   bool   `json:"test"`
	Url    string `json:"url,omitempty"`
	Locale string `json:"locale"`
	Images []struct {
		Ratio    string `json:"ratio"`
		Url      string `json:"url"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Fallback bool   `json:"fallback"`
	} `json:"images"`
	Sales struct {
		Public struct {
			StartDateTime time.Time `json:"startDateTime"`
			StartTBD      bool      `json:"startTBD"`
			StartTBA      bool      `json:"startTBA"`
			EndDateTime   time.Time `json:"endDateTime"`
		} `json:"public"`
		Presales []struct {
			StartDateTime time.Time `json:"startDateTime"`
			EndDateTime   time.Time `json:"endDateTime"`
			Name          string    `json:"name"`
		} `json:"presales"`
	} `json:"sales,omitempty"`
	Dates struct {
		Start struct {
			LocalDate      string    `json:"localDate"`
			LocalTime      string    `json:"localTime"`
			DateTime       time.Time `json:"dateTime"`
			DateTBD        bool      `json:"dateTBD"`
			DateTBA        bool      `json:"dateTBA"`
			TimeTBA        bool      `json:"timeTBA"`
			NoSpecificTime bool      `json:"noSpecificTime"`
		} `json:"start"`
		Timezone string `json:"timezone"`
		Status   struct {
			Code string `json:"code"`
		} `json:"status"`
		SpanMultipleDays bool `json:"spanMultipleDays"`
		End              struct {
			LocalDate      string    `json:"localDate"`
			LocalTime      string    `json:"localTime"`
			DateTime       time.Time `json:"dateTime"`
			Approximate    bool      `json:"approximate"`
			NoSpecificTime bool      `json:"noSpecificTime"`
		} `json:"end,omitempty"`
	} `json:"dates"`
	Classifications []struct {
		Primary bool `json:"primary"`
		Segment struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"segment"`
		Genre struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"genre"`
		SubGenre struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"subGenre"`
		Type struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"type"`
		SubType struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"subType"`
		Family bool `json:"family"`
	} `json:"classifications,omitempty"`
	Promoter struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"promoter,omitempty"`
	Promoters []struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"promoters,omitempty"`
	Info        string `json:"info,omitempty"`
	PleaseNote  string `json:"pleaseNote,omitempty"`
	PriceRanges []struct {
		Type     string  `json:"type"`
		Currency string  `json:"currency"`
		Min      float64 `json:"min"`
		Max      float64 `json:"max"`
	} `json:"priceRanges,omitempty"`
	Seatmap struct {
		StaticUrl string `json:"staticUrl"`
		Id        string `json:"id"`
	} `json:"seatmap,omitempty"`
	Accessibility struct {
		TicketLimit int    `json:"ticketLimit"`
		Id          string `json:"id"`
	} `json:"accessibility,omitempty"`
	TicketLimit struct {
		Info string `json:"info"`
		Id   string `json:"id"`
	} `json:"ticketLimit,omitempty"`
	AgeRestrictions struct {
		LegalAgeEnforced bool   `json:"legalAgeEnforced"`
		Id               string `json:"id"`
	} `json:"ageRestrictions,omitempty"`
	Ticketing struct {
		SafeTix struct {
			Enabled          bool `json:"enabled"`
			InAppOnlyEnabled bool `json:"inAppOnlyEnabled,omitempty"`
		} `json:"safeTix"`
		AllInclusivePricing struct {
			Enabled bool `json:"enabled"`
		} `json:"allInclusivePricing,omitempty"`
		Id string `json:"id"`
	} `json:"ticketing"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Attractions []struct {
			Href string `json:"href"`
		} `json:"attractions,omitempty"`
		Venues []struct {
			Href string `json:"href"`
		} `json:"venues"`
	} `json:"_links"`
	Embedded struct {
		Venues []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Id         string `json:"id"`
			Test       bool   `json:"test"`
			Url        string `json:"url,omitempty"`
			Locale     string `json:"locale"`
			PostalCode string `json:"postalCode"`
			Timezone   string `json:"timezone"`
			City       struct {
				Name string `json:"name"`
			} `json:"city"`
			State struct {
				Name      string `json:"name"`
				StateCode string `json:"stateCode"`
			} `json:"state"`
			Country struct {
				Name        string `json:"name"`
				CountryCode string `json:"countryCode"`
			} `json:"country"`
			Address struct {
				Line1 string `json:"line1"`
			} `json:"address"`
			Location struct {
				Longitude string `json:"longitude"`
				Latitude  string `json:"latitude"`
			} `json:"location"`
			Markets []struct {
				Name string `json:"name"`
				Id   string `json:"id"`
			} `json:"markets,omitempty"`
			Dmas []struct {
				Id int `json:"id"`
			} `json:"dmas,omitempty"`
			BoxOfficeInfo struct {
				PhoneNumberDetail     string `json:"phoneNumberDetail"`
				OpenHoursDetail       string `json:"openHoursDetail"`
				AcceptedPaymentDetail string `json:"acceptedPaymentDetail"`
				WillCallDetail        string `json:"willCallDetail"`
			} `json:"boxOfficeInfo,omitempty"`
			ParkingDetail string `json:"parkingDetail,omitempty"`
			GeneralInfo   struct {
				GeneralRule string `json:"generalRule"`
				ChildRule   string `json:"childRule"`
			} `json:"generalInfo,omitempty"`
			UpcomingEvents struct {
				Archtics     int `json:"archtics"`
				Tmr          int `json:"tmr,omitempty"`
				Ticketmaster int `json:"ticketmaster"`
				Total        int `json:"_total"`
				Filtered     int `json:"_filtered"`
			} `json:"upcomingEvents"`
			Ada struct {
				AdaPhones     string `json:"adaPhones"`
				AdaCustomCopy string `json:"adaCustomCopy"`
				AdaHours      string `json:"adaHours"`
			} `json:"ada,omitempty"`
			Links struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
		} `json:"venues"`
		Attractions []struct {
			Name          string `json:"name"`
			Type          string `json:"type"`
			Id            string `json:"id"`
			Test          bool   `json:"test"`
			Url           string `json:"url"`
			Locale        string `json:"locale"`
			ExternalLinks struct {
				Twitter []struct {
					Url string `json:"url"`
				} `json:"twitter"`
				Facebook []struct {
					Url string `json:"url"`
				} `json:"facebook"`
				Wiki []struct {
					Url string `json:"url"`
				} `json:"wiki"`
				Instagram []struct {
					Url string `json:"url"`
				} `json:"instagram"`
				Homepage []struct {
					Url string `json:"url"`
				} `json:"homepage"`
			} `json:"externalLinks"`
			Aliases []string `json:"aliases,omitempty"`
			Images  []struct {
				Ratio       string `json:"ratio"`
				Url         string `json:"url"`
				Width       int    `json:"width"`
				Height      int    `json:"height"`
				Fallback    bool   `json:"fallback"`
				Attribution string `json:"attribution,omitempty"`
			} `json:"images"`
			Classifications []struct {
				Primary bool `json:"primary"`
				Segment struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"segment"`
				Genre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"genre"`
				SubGenre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subGenre"`
				Type struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"type"`
				SubType struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subType"`
				Family bool `json:"family"`
			} `json:"classifications"`
			UpcomingEvents struct {
				Tmr          int `json:"tmr"`
				Ticketmaster int `json:"ticketmaster"`
				Total        int `json:"_total"`
				Filtered     int `json:"_filtered"`
			} `json:"upcomingEvents"`
			Links struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
		} `json:"attractions,omitempty"`
	} `json:"_embedded"`
}

const (
	AttractionIDKraken   = "K8vZ917_vgV"
	AttractionIDSeahawks = "K8vZ9171oU7"
	AttractionIDMariners = "K8vZ9171o6f"

	TicketmasterEventSearchAPI   = "https://app.ticketmaster.com/discovery/v2/events"
	TicketmasterApiKeySecretName = "TICKETMASTER_API_KEY_SECRET_NAME"
)

// seattleVenueMap is a map of venues to ticketmaster's internal venue ID for venues we should look at
var seattleVenueMap = map[string]string{
	"Climate Pledge Arena": "KovZ917Ahkk",
	"Lumen Field":          "KovZpZAEknnA",
	"T-Mobile Park":        "KovZpZAEevAA",
	"WAMU Theater":         "KovZpZAFFE7A",
}

// seattleTeamsMap is a set of attraction IDs (read: sports teams) that we want to ignore for this so we don't
// count these events twice (since we check for them in other places). Note for later ... should we just do _everything_
// via the tikcetmaster API? probably
var seattleTeamsMap = map[string]bool{
	AttractionIDKraken:   true,
	AttractionIDSeahawks: true,
	AttractionIDMariners: true,
}

func beginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func beginningOfTomorrow(t time.Time) time.Time {
	tomorrow := t.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, t.Location())
}

func eventShouldBeIgnored(e *TicketmasterEvent) bool {
	if e.Classifications == nil || len(e.Classifications) == 0 {
		return true
	}

	if len(e.Embedded.Attractions) == 0 {
		return true
	}

	for _, curr := range e.Embedded.Attractions {
		if seattleTeamsMap[curr.Id] {
			return true
		}
	}

	return false
}

func getEventForVenueID(ctx context.Context, apiKey string, venueName string, venueID string, searchStart string, searchEnd string) (*Event, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TicketmasterEventSearchAPI, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("venueId", venueID)
	q.Add("apikey", apiKey)
	q.Add("startDateTime", searchStart)
	q.Add("endDateTime", searchEnd)
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Str("status", resp.Status).Msg("could not read error response body")
			return nil, fmt.Errorf("events: getEventForVenueID: could not read error body: %w", err)
		}
		log.Error().Str("status", resp.Status).Msg("error retrieving data from ticketmaster")
		return nil, fmt.Errorf("events: getEventForVenueID: could not retireve data from ticketmaster: %s", string(body))
	}

	var payload TicketmasterEventSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		return nil, err
	}

	// TODO: make this return multiple events if needed
	for _, e := range payload.Embedded.Events {

		if eventShouldBeIgnored(&e) {
			log.Info().Str("venue", venueName).Str("event_name", e.Name).Msg("ignoring event")
			continue
		}

		startTime := e.Dates.Start.DateTime.In(SeattleTimeZone)

		return &Event{
			RawDescription: fmt.Sprintf("%s is at %s. It starts at %s", e.Name, venueName, startTime.Format(localTimeDateFormat)),
			RawTime:        e.Dates.Start.DateTime.Unix(),
		}, nil
	}

	return nil, nil

}

func GetOtherEvents(ctx context.Context) ([]*Event, error) {
	ticketmasterApiKeySecretName := os.Getenv(TicketmasterApiKeySecretName)
	if ticketmasterApiKeySecretName == "" {
		log.Warn().Str("env_var_name", TicketmasterApiKeySecretName).Msg("environment variable not set. Not querying ticketmaster")
		return nil, nil
	}

	apiKey, err := secrets.GetSecretString(ctx, ticketmasterApiKeySecretName)
	if err != nil {
		return nil, fmt.Errorf("events: GetOtherEvents: could not get ticketmaster secret: %w", err)
	}

	today := time.Now().In(SeattleTimeZone)

	start := beginningOfDay(today).Format(time.RFC3339)
	end := beginningOfTomorrow(today).Format(time.RFC3339)

	var res []*Event

	for venueName, venueID := range seattleVenueMap {
		var e *Event
		e, err = getEventForVenueID(ctx, apiKey, venueName, venueID, start, end)
		if err != nil {
			return nil, fmt.Errorf("events: GetOtherEvents: could not query for ticketmaster data: %w", err)
		}
		if e != nil {
			res = append(res, e)
		}

		// ticketmaster limits us to 5 calls per second -- add a delay so we don't blow through that
		time.Sleep(300 * time.Millisecond)
	}

	return res, nil
}

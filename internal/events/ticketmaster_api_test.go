package events

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

//go:embed sample_tm_payloads/simple.json
var sampleResponseText []byte

//go:embed sample_tm_payloads/tba.json
var tbaResponseText []byte

func TestTicketmasterFetcher_GetEvents(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		var sampleResponse TicketmasterEventSearchResponse
		err := json.Unmarshal(sampleResponseText, &sampleResponse)
		require.NoError(t, err)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			require.NoError(t, err)

			assert.Equal(t, "test-api-key", r.FormValue("apikey"))
			assert.Equal(t, "CPA-VENUE-ID", r.FormValue("venueId"))

			parsedStart, err := time.Parse(time.RFC3339, r.FormValue("startDateTime"))
			require.NoError(t, err)
			parsedEnd, err := time.Parse(time.RFC3339, r.FormValue("endDateTime"))
			require.NoError(t, err)

			today := time.Now().In(SeattleTimeZone)

			assert.Equal(t, 0, parsedStart.Hour())
			assert.Equal(t, 0, parsedStart.Minute())
			assert.Equal(t, 0, parsedStart.Second())
			assert.Equal(t, 0, parsedEnd.Hour())
			assert.Equal(t, 0, parsedEnd.Minute())
			assert.Equal(t, 0, parsedEnd.Second())
			assert.Equal(t, today.Year(), parsedStart.Year())
			assert.Equal(t, today.AddDate(0, 0, 2).Year(), parsedEnd.Year())
			assert.Equal(t, today.YearDay(), parsedStart.YearDay())
			assert.Equal(t, today.AddDate(0, 0, 2).YearDay(), parsedEnd.YearDay())

			startTime := time.Date(today.Year(), today.Month(), today.Day(), 19, 0, 0, 0, SeattleTimeZone)

			for _, curr := range sampleResponse.Embedded.Events {
				curr.Dates.Start.LocalDate = today.Format("2006-01-02")
				curr.Dates.Start.DateTime = startTime
			}

			rebuildReponse, err := json.Marshal(sampleResponse)
			require.NoError(t, err)
			w.Write(rebuildReponse)
		}))
		f := &ticketmasterFetcher{
			venues: map[string]string{
				"Climate Pledge Arena": "CPA-VENUE-ID",
			},
			attractionIDs: seattleTeamAttractionIDs,
			limiter:       rate.NewLimiter(10, 1),
			apiKey:        "test-api-key",
			baseURL:       srv.URL,
		}

		today, tomorrow, err := f.GetEvents(context.TODO())
		require.NoError(t, err)

		require.Len(t, today, 1)
		assert.Len(t, tomorrow, 0)

		returnedEvent := today[0]

		assert.Equal(t, "Climate Pledge Arena", returnedEvent.Venue)
		assert.Equal(t, "Vegas Golden Knights", returnedEvent.Opponent)
		assert.Equal(t, "7:00 PM", returnedEvent.LocalTime)
	})

	t.Run("time tba event", func(t *testing.T) {
		var sampleResponse TicketmasterEventSearchResponse
		err := json.Unmarshal(tbaResponseText, &sampleResponse)
		require.NoError(t, err)

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			require.NoError(t, err)

			today := time.Now().In(SeattleTimeZone)

			for _, curr := range sampleResponse.Embedded.Events {
				curr.Dates.Start.LocalDate = today.Format("2006-01-02")
			}

			rebuildReponse, err := json.Marshal(sampleResponse)
			require.NoError(t, err)
			w.Write(rebuildReponse)
		}))
		f := &ticketmasterFetcher{
			venues: map[string]string{
				"Climate Pledge Arena": "CPA-VENUE-ID",
			},
			attractionIDs: seattleTeamAttractionIDs,
			limiter:       rate.NewLimiter(10, 1),
			apiKey:        "test-api-key",
			baseURL:       srv.URL,
		}

		today, tomorrow, err := f.GetEvents(context.TODO())
		require.NoError(t, err)

		require.Len(t, today, 1)
		assert.Len(t, tomorrow, 0)

		returnedEvent := today[0]

		assert.Equal(t, "Climate Pledge Arena", returnedEvent.Venue)
		assert.Equal(t, "Vegas Golden Knights", returnedEvent.Opponent)
		assert.Equal(t, "TBA", returnedEvent.LocalTime)
	})
}

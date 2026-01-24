package events

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

//go:embed testdata/*.json
var testData embed.FS

type test struct {
	name          string
	file          string
	date          time.Time
	checkToday    func(*testing.T, []*Event)
	checkTomorrow func(*testing.T, []*Event)
}

func TestTicketmasterFetcher_GetEvents(t *testing.T) {
	tests := []test{
		{
			name: "Basic test",
			file: "simple.json",
			date: time.Date(2026, time.February, 14, 0, 0, 0, 0, SeattleTimeZone),
			checkToday: func(t *testing.T, events []*Event) {
				assert.Len(t, events, 1)
				returnedEvent := events[0]
				assert.Equal(t, "Jo Koy: Just Being Koy Tour is at Climate Pledge Arena. It starts at 8:00 PM", returnedEvent.RawDescription)
			},
			checkTomorrow: func(t *testing.T, events []*Event) {
				assert.Len(t, events, 1)
				returnedEvent := events[0]
				assert.Equal(t, "GHOST: Skeletour World Tour 2026 is at Climate Pledge Arena. It starts at 8:00 PM", returnedEvent.RawDescription)
			},
		},
		{
			name: "Interesting sports",
			file: "battle_of_sound_harlem_globetrotters.json",
			date: time.Date(2026, time.January, 31, 0, 0, 0, 0, SeattleTimeZone),
			checkToday: func(t *testing.T, events []*Event) {
				assert.Len(t, events, 1)

				assert.Equal(t, "Battle of the Sound: Seattle Thunderbirds vs Everett Silvertips is at Climate Pledge Arena. It starts at 6:05 PM", events[0].RawDescription)
			},
			checkTomorrow: func(t *testing.T, events []*Event) {
				assert.Len(t, events, 1)

				assert.Equal(t, "The Harlem Globetrotters 100 Year Tour is at Climate Pledge Arena. It starts at 3:00 PM", events[0].RawDescription)
			},
		},
		{
			name: "Ignore these events",
			file: "ignore_these_events.json",
			date: time.Date(2026, time.January, 18, 0, 0, 0, 0, SeattleTimeZone),
			checkToday: func(t *testing.T, events []*Event) {
				assert.Empty(t, events)
			},
			checkTomorrow: func(t *testing.T, events []*Event) {
				assert.Empty(t, events)
			},
		},
		{
			name: "Ignore fanfest",
			file: "ignore_fanfest.json",
			date: time.Date(2026, time.January, 31, 0, 0, 0, 0, SeattleTimeZone),
			checkToday: func(t *testing.T, events []*Event) {
				assert.Empty(t, events)
			},
			checkTomorrow: func(t *testing.T, events []*Event) {
				assert.Empty(t, events)
			},
		},
	}

	for _, curr := range tests {
		t.Run(curr.name, func(t *testing.T) {
			subFS, err := fs.Sub(testData, "testdata")
			require.NoError(t, err)

			output, err := fs.ReadFile(subFS, curr.file)
			require.NoError(t, err)

			var sampleResponse TicketmasterEventSearchResponse
			err = json.Unmarshal(output, &sampleResponse)
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

				assert.Equal(t, 0, parsedStart.Hour())
				assert.Equal(t, 0, parsedStart.Minute())
				assert.Equal(t, 0, parsedStart.Second())
				assert.Equal(t, 0, parsedEnd.Hour())
				assert.Equal(t, 0, parsedEnd.Minute())
				assert.Equal(t, 0, parsedEnd.Second())
				assert.Equal(t, curr.date.Year(), parsedStart.Year())
				assert.Equal(t, curr.date.AddDate(0, 0, 2).Year(), parsedEnd.Year())
				assert.Equal(t, curr.date.YearDay(), parsedStart.YearDay())
				assert.Equal(t, curr.date.AddDate(0, 0, 2).YearDay(), parsedEnd.YearDay())

				_, _ = w.Write(output)
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

			today, tomorrow, err := f.GetEvents(context.TODO(), curr.date, curr.date.AddDate(0, 0, 1))
			require.NoError(t, err)

			curr.checkToday(t, today)
			curr.checkTomorrow(t, tomorrow)
		})
	}
}

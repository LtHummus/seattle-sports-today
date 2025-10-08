package renderjson

import (
	"encoding/json"
	"fmt"

	"github.com/lthummus/seattle-sports-today/internal/events"
)

func renderEventSlice(x []*events.Event) []map[string]string {
	renderableEvents := make([]map[string]string, len(x))

	for i, curr := range x {
		e := map[string]string{}
		e["description"] = curr.String()
		if curr.Venue != "" {
			e["venue"] = curr.Venue
		}
		if curr.TeamName != "" {
			e["team_name"] = curr.TeamName
		}
		if curr.Opponent != "" {
			e["opponent"] = curr.Opponent
		}
		if curr.LocalTime != "" {
			e["local_time"] = curr.LocalTime
		}
		renderableEvents[i] = e
	}

	return renderableEvents
}

func RenderJSON(results *events.EventResults) ([]byte, error) {
	data := map[string]any{
		"date":            events.SeattleToday.Format("2006-01-02"),
		"events":          renderEventSlice(results.TodayEvent),
		"tomorrow_events": renderEventSlice(results.TomorrowEvents),
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("renderJSON: could not render: %w", err)
	}

	return payload, nil
}

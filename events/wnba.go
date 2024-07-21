package events

import (
	"context"
)

func GetStormGame(ctx context.Context) (*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/basketball/wnba/scoreboard", "Seattle Storm", "SEA")
}

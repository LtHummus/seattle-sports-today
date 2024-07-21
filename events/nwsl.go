package events

import (
	"context"
)

func GetReignGame(ctx context.Context) (*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/soccer/usa.nwsl/scoreboard", "Seattle Reign", "SEA")
}

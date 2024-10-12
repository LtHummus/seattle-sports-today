package events

import (
	"context"
)

func GetSeahawksGame(ctx context.Context) ([]*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard", "Seattle Seahawks", "SEA")
}

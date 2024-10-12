package events

import (
	"context"
)

func GetSoundersGame(ctx context.Context) ([]*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/soccer/usa.1/scoreboard", "Seattle Sounders", "SEA")
}

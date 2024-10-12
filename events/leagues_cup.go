package events

import "context"

func GetSoundersLeagueCupGame(ctx context.Context) ([]*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/soccer/concacaf.leagues.cup/scoreboard", "Seattle Sounders", "SEA")
}

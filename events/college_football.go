package events

import "context"

func GetHuskiesFootballGame(ctx context.Context) (*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/football/college-football/scoreboard", "Washington Huskies", "WASH")
}

package events

import (
	"context"
)

func GetHuskiesBasketballMensGame(ctx context.Context) ([]*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/basketball/mens-college-basketball/scoreboard", "Washington Huskies (Men's Basketball)", "WASH")
}

func GetHuskiesBasketballWomensGame(ctx context.Context) ([]*Event, error) {
	return queryESPN(ctx, "https://site.api.espn.com/apis/site/v2/sports/basketball/womens-college-basketball/scoreboard", "Washington Huskies (Women's Basketball)", "WASH")
}

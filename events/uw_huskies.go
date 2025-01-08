package events

import "context"

const (
	huskiesFootballURL         = "https://site.api.espn.com/apis/site/v2/sports/football/college-football/teams/WASH"
	huskiesMensBasketballURL   = "https://site.api.espn.com/apis/site/v2/sports/basketball/mens-college-basketball/teams/264"
	huskiesWomensBasketballURL = "https://site.api.espn.com/apis/site/v2/sports/basketball/womens-college-basketball/teams/264"

	huskiesFootballName         = "Washington Huskies (Football)"
	huskiesMensBasketballName   = "Washington Huskies (Men's Basketball)"
	huskiesWomensBasketballName = "Washington Huskies (Women's Basketball)"

	huskyStadium        = "Husky Stadium"
	alaskaAirlinesArena = "Alaska Airlines Arena"
)

func GetHuskiesBasketballMensGame(ctx context.Context) ([]*Event, error) {
	return queryESPNTeam(ctx, huskiesMensBasketballURL, huskiesMensBasketballName, alaskaAirlinesArena)
}

func GetHuskiesBasketballWomensGame(ctx context.Context) ([]*Event, error) {
	return queryESPNTeam(ctx, huskiesWomensBasketballURL, huskiesWomensBasketballName, alaskaAirlinesArena)
}

func GetHuskiesFootballGame(ctx context.Context) ([]*Event, error) {
	return queryESPNTeam(ctx, huskiesFootballURL, huskiesFootballName, huskyStadium)
}

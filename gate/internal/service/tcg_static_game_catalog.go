package service

import (
	"strconv"
	"strings"
)

type staticTCGGameCatalogItem struct {
	ProductType   int
	GameType      string
	GameName      string
	TCGGameCode   string
	ProductCode   string
	Platform      string
	GameSubType   string
	ShowIcon      string
	TrialSupport  bool
	DisplayStatus int
}

var staticTCGGameCatalog = []staticTCGGameCatalogItem{
	// From TCG TW docs sample `tgl` response.
	{ProductType: 7, GameType: "RNG", GameName: "Big fish eat small fish_大鱼吃小鱼", TCGGameCode: "G00005", ProductCode: "GG", Platform: "flash,html5", GameSubType: "SM", TrialSupport: true, DisplayStatus: 0},
	// From Launch Game API examples in docs.
	{ProductType: 4, GameType: "RNG", GameName: "AG Sample Game", TCGGameCode: "A00070", ProductCode: "AG", Platform: "html5-desktop,html5", GameSubType: "RNG", TrialSupport: true, DisplayStatus: 0},
	{ProductType: 2, GameType: "LOTT", GameName: "TCG Lottery Lobby", TCGGameCode: "Lobby", ProductCode: "TCG LOTTO", Platform: "html5-desktop", GameSubType: "LOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 2, GameType: "LOTT", GameName: "TCG Lottery Mobile Game List", TCGGameCode: "Game_List", ProductCode: "TCG LOTTO", Platform: "html5", GameSubType: "LOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 2, GameType: "LOTT", GameName: "TCG Lottery Sample Single Game", TCGGameCode: "4DTWC@", ProductCode: "TCG LOTTO", Platform: "html5", GameSubType: "LOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 2, GameType: "ELOTT", GameName: "TCG eLottery Lobby", TCGGameCode: "Lobby", ProductCode: "TCG LOTTO", Platform: "html5-desktop", GameSubType: "ELOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 2, GameType: "ELOTT", GameName: "TCG eLottery Mobile Game List", TCGGameCode: "game_list", ProductCode: "TCG LOTTO", Platform: "html5", GameSubType: "ELOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 384, GameType: "ELOTT", GameName: "TCG SEA Lobby", TCGGameCode: "lobby", ProductCode: "TCGSVNC", Platform: "flash", GameSubType: "ELOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 420, GameType: "LOTT", GameName: "TCG VN Lotto Lobby", TCGGameCode: "lobby", ProductCode: "TCGVNLOTT", Platform: "WEB", GameSubType: "LOTT", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 420, GameType: "ELOTT", GameName: "TCG VN eLotto Lobby", TCGGameCode: "lobby", ProductCode: "TCGVNLOTT", Platform: "WEB", GameSubType: "ELOTT", TrialSupport: false, DisplayStatus: 0},
	// From TCG Live examples in docs (`gml` / `gtl` sections).
	{ProductType: 460, GameType: "LIVE", GameName: "TCG Live Lobby", TCGGameCode: "lobby", ProductCode: "TCGLIVE", Platform: "MOBILE", GameSubType: "LIVE", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 460, GameType: "LIVE", GameName: "百人牛牛", TCGGameCode: "TCGNNP", ProductCode: "TCGLIVE", Platform: "MOBILE", GameSubType: "NN", TrialSupport: false, DisplayStatus: 0},
	{ProductType: 460, GameType: "LIVE", GameName: "TCG Live K3 Room", TCGGameCode: "MILK3", ProductCode: "TCGLIVE", Platform: "MOBILE", GameSubType: "SIC", TrialSupport: false, DisplayStatus: 0},
}

func buildStaticTCGGameCatalog(gameTypeProducts map[string][]int) []TCGGameListItem {
	items := make([]TCGGameListItem, 0)
	requestedGameTypes := make(map[string]struct{}, len(gameTypeProducts))
	productTypeGameTypeSet := make(map[string]struct{})
	for gameType, productTypes := range gameTypeProducts {
		normalizedGameType := strings.ToUpper(strings.TrimSpace(gameType))
		if normalizedGameType == "" {
			continue
		}
		requestedGameTypes[normalizedGameType] = struct{}{}
		for _, productType := range productTypes {
			productTypeGameTypeSet[strconv.Itoa(productType)+":"+normalizedGameType] = struct{}{}
		}
	}

	for _, item := range staticTCGGameCatalog {
		gameType := strings.ToUpper(strings.TrimSpace(item.GameType))
		if len(requestedGameTypes) > 0 {
			if _, ok := requestedGameTypes[gameType]; !ok {
				continue
			}
		}
		if len(productTypeGameTypeSet) > 0 {
			if _, ok := productTypeGameTypeSet[strconv.Itoa(item.ProductType)+":"+gameType]; !ok {
				continue
			}
		}

		items = append(items, TCGGameListItem{
			DisplayStatus: item.DisplayStatus,
			GameType:      item.GameType,
			GameName:      item.GameName,
			TCGGameCode:   item.TCGGameCode,
			ProductCode:   item.ProductCode,
			ProductType:   intToString(item.ProductType),
			Platform:      item.Platform,
			GameSubType:   item.GameSubType,
			ShowIcon:      item.ShowIcon,
			TrialSupport:  item.TrialSupport,
		})
	}

	return items
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

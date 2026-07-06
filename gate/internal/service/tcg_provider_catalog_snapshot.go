package service

import (
	"sort"
	"strconv"
	"strings"
)

const (
	tcgCatalogItemKindLaunch = "launch"
	tcgCatalogItemKindGroup  = "group"
)

type TCGProviderCatalogCategory struct {
	Key   string                           `json:"key"`
	Label string                           `json:"label"`
	Items []TCGProviderCatalogCategoryItem `json:"items"`
}

type TCGProviderCatalogCategoryItem struct {
	Kind             string            `json:"kind"`
	DisplayStatus    int               `json:"display_status"`
	GameType         string            `json:"game_type"`
	GameName         string            `json:"game_name"`
	TCGGameCode      string            `json:"tcg_game_code"`
	ProductCode      string            `json:"product_code"`
	ProductType      int               `json:"product_type"`
	ProductTypeValue string            `json:"product_type_value"`
	Platform         string            `json:"platform"`
	GameSubType      string            `json:"game_sub_type"`
	ShowIcon         string            `json:"show_icon,omitempty"`
	TrialSupport     bool              `json:"trial_support"`
	ChildCount       int               `json:"child_count,omitempty"`
	Children         []TCGGameListItem `json:"children,omitempty"`
}

type tcgCatalogCategoryMeta struct {
	Key   string
	Label string
}

var (
	tcgCockfightProductTypes = map[string]struct{}{
		"132": {},
		"202": {},
	}
	tcgCatalogCategoryOrder = []tcgCatalogCategoryMeta{
		{Key: "lottery", Label: "Xo so"},
		{Key: "casino", Label: "Casino"},
		{Key: "slots", Label: "No hu"},
		{Key: "fish", Label: "Ban ca"},
		{Key: "sports", Label: "The thao"},
		{Key: "cards", Label: "Game bai"},
		{Key: "cockfight", Label: "Da ga"},
	}
)

func buildTCGProviderCatalogCategories(games []TCGGameListItem) []TCGProviderCatalogCategory {
	categoryItems := make(map[string][]TCGGameListItem)
	categoryMetaByKey := make(map[string]tcgCatalogCategoryMeta)

	for _, item := range dedupeTCGGameListItems(games) {
		meta, ok := resolveTCGCatalogCategory(item)
		if !ok {
			continue
		}

		categoryItems[meta.Key] = append(categoryItems[meta.Key], item)
		categoryMetaByKey[meta.Key] = meta
	}

	categories := make([]TCGProviderCatalogCategory, 0, len(categoryItems))
	seen := make(map[string]struct{}, len(categoryItems))
	for _, meta := range tcgCatalogCategoryOrder {
		items := buildTCGCategoryCatalogItems(meta.Key, categoryItems[meta.Key])
		if len(items) == 0 {
			continue
		}

		categories = append(categories, TCGProviderCatalogCategory{
			Key:   meta.Key,
			Label: meta.Label,
			Items: items,
		})
		seen[meta.Key] = struct{}{}
	}

	for key, items := range categoryItems {
		if _, ok := seen[key]; ok {
			continue
		}
		if len(items) == 0 {
			continue
		}

		meta := categoryMetaByKey[key]
		categories = append(categories, TCGProviderCatalogCategory{
			Key:   meta.Key,
			Label: meta.Label,
			Items: buildTCGCategoryCatalogItems(meta.Key, items),
		})
	}

	return categories
}

func buildTCGCategoryCatalogItems(categoryKey string, items []TCGGameListItem) []TCGProviderCatalogCategoryItem {
	switch strings.TrimSpace(categoryKey) {
	case "casino":
		return buildCasinoCatalogItems(items)
	case "fish":
		return buildFishCatalogItems(items)
	case "sports":
		return buildSportsCatalogItems(items)
	default:
		return buildDirectCatalogItems(items)
	}
}

func buildCasinoCatalogItems(items []TCGGameListItem) []TCGProviderCatalogCategoryItem {
	grouped := groupTCGItemsByProductType(items)
	productKeys := orderedTCGProductKeys(grouped)
	result := make([]TCGProviderCatalogCategoryItem, 0, len(productKeys))

	for _, productKey := range productKeys {
		children := normalizeTCGCatalogChildren(grouped[productKey])
		if len(children) == 0 {
			continue
		}

		hero := pickCasinoCatalogHero(children)
		result = append(result, buildGroupedCatalogItem(hero, children))
	}

	sortCatalogItems(result)
	return result
}

func buildSportsCatalogItems(items []TCGGameListItem) []TCGProviderCatalogCategoryItem {
	grouped := groupTCGItemsByProductType(items)
	productKeys := orderedTCGProductKeys(grouped)
	result := make([]TCGProviderCatalogCategoryItem, 0, len(productKeys))

	for _, productKey := range productKeys {
		candidates := normalizeTCGCatalogChildren(grouped[productKey])
		if len(candidates) == 0 {
			continue
		}

		launchItem := pickSportsCatalogLobby(candidates)
		result = append(result, buildLaunchCatalogItem(launchItem))
	}

	sortCatalogItems(result)
	return result
}

func buildFishCatalogItems(items []TCGGameListItem) []TCGProviderCatalogCategoryItem {
	grouped := groupTCGItemsByProductType(items)
	productKeys := orderedTCGProductKeys(grouped)
	result := make([]TCGProviderCatalogCategoryItem, 0, len(productKeys))

	for _, productKey := range productKeys {
		children := normalizeTCGCatalogChildren(grouped[productKey])
		if len(children) == 0 {
			continue
		}

		hero := pickFishCatalogHero(children)
		result = append(result, buildGroupedCatalogItem(hero, children))
	}

	sortCatalogItems(result)
	return result
}

func buildDirectCatalogItems(items []TCGGameListItem) []TCGProviderCatalogCategoryItem {
	normalized := normalizeTCGCatalogChildren(items)
	result := make([]TCGProviderCatalogCategoryItem, 0, len(normalized))
	for _, item := range normalized {
		result = append(result, buildLaunchCatalogItem(item))
	}

	sortCatalogItems(result)
	return result
}

func buildLaunchCatalogItem(item TCGGameListItem) TCGProviderCatalogCategoryItem {
	return TCGProviderCatalogCategoryItem{
		Kind:             tcgCatalogItemKindLaunch,
		DisplayStatus:    item.DisplayStatus,
		GameType:         strings.TrimSpace(item.GameType),
		GameName:         strings.TrimSpace(item.GameName),
		TCGGameCode:      strings.TrimSpace(item.TCGGameCode),
		ProductCode:      strings.TrimSpace(item.ProductCode),
		ProductType:      parseTCGProductType(item.ProductType),
		ProductTypeValue: strings.TrimSpace(item.ProductType),
		Platform:         strings.TrimSpace(item.Platform),
		GameSubType:      strings.TrimSpace(item.GameSubType),
		ShowIcon:         strings.TrimSpace(item.ShowIcon),
		TrialSupport:     item.TrialSupport,
	}
}

func buildGroupedCatalogItem(hero TCGGameListItem, children []TCGGameListItem) TCGProviderCatalogCategoryItem {
	groupItem := buildLaunchCatalogItem(hero)
	groupItem.Kind = tcgCatalogItemKindGroup
	groupItem.ChildCount = len(children)
	groupItem.Children = append([]TCGGameListItem(nil), children...)
	return groupItem
}

func resolveTCGCatalogCategory(item TCGGameListItem) (tcgCatalogCategoryMeta, bool) {
	gameType := strings.ToUpper(strings.TrimSpace(item.GameType))
	productType := strings.TrimSpace(item.ProductType)

	switch gameType {
	case "LIVE":
		return tcgCatalogCategoryMeta{Key: "casino", Label: "Casino"}, true
	case "RNG":
		return tcgCatalogCategoryMeta{Key: "slots", Label: "No hu"}, true
	case "FISH":
		return tcgCatalogCategoryMeta{Key: "fish", Label: "Ban ca"}, true
	case "PVP":
		return tcgCatalogCategoryMeta{Key: "cards", Label: "Game bai"}, true
	case "LOTT", "ELOTT":
		return tcgCatalogCategoryMeta{Key: "lottery", Label: "Xo so"}, true
	case "SPORT", "SPORTS":
		if _, ok := tcgCockfightProductTypes[productType]; ok {
			return tcgCatalogCategoryMeta{Key: "cockfight", Label: "Da ga"}, true
		}
		return tcgCatalogCategoryMeta{Key: "sports", Label: "The thao"}, true
	default:
		return tcgCatalogCategoryMeta{}, false
	}
}

func dedupeTCGGameListItems(items []TCGGameListItem) []TCGGameListItem {
	seen := make(map[string]struct{}, len(items))
	result := make([]TCGGameListItem, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.ProductType) + ":" + strings.TrimSpace(item.TCGGameCode)
		if key == ":" {
			key = strings.TrimSpace(item.GameType) + ":" + strings.TrimSpace(item.GameName)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func groupTCGItemsByProductType(items []TCGGameListItem) map[string][]TCGGameListItem {
	grouped := make(map[string][]TCGGameListItem)
	for _, item := range items {
		key := strings.TrimSpace(item.ProductType)
		if key == "" {
			key = "0"
		}
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}

func orderedTCGProductKeys(grouped map[string][]TCGGameListItem) []string {
	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		leftInt := parseTCGProductType(keys[i])
		rightInt := parseTCGProductType(keys[j])
		if leftInt != rightInt {
			return leftInt < rightInt
		}
		return keys[i] < keys[j]
	})
	return keys
}

func normalizeTCGCatalogChildren(items []TCGGameListItem) []TCGGameListItem {
	normalized := dedupeTCGGameListItems(items)
	sort.SliceStable(normalized, func(i, j int) bool {
		left := normalized[i]
		right := normalized[j]
		if left.DisplayStatus != right.DisplayStatus {
			return left.DisplayStatus < right.DisplayStatus
		}
		leftHasIcon := strings.TrimSpace(left.ShowIcon) != ""
		rightHasIcon := strings.TrimSpace(right.ShowIcon) != ""
		if leftHasIcon != rightHasIcon {
			return leftHasIcon
		}
		if left.ProductType != right.ProductType {
			return parseTCGProductType(left.ProductType) < parseTCGProductType(right.ProductType)
		}
		leftName := strings.TrimSpace(left.GameName)
		rightName := strings.TrimSpace(right.GameName)
		if leftName != rightName {
			return leftName < rightName
		}
		return strings.TrimSpace(left.TCGGameCode) < strings.TrimSpace(right.TCGGameCode)
	})
	return normalized
}

func pickCasinoCatalogHero(items []TCGGameListItem) TCGGameListItem {
	best := items[0]
	bestRank := casinoCatalogHeroRank(best)
	for _, item := range items[1:] {
		rank := casinoCatalogHeroRank(item)
		if rank > bestRank {
			best = item
			bestRank = rank
			continue
		}
		if rank == bestRank && compareCatalogNames(item.GameName, best.GameName) < 0 {
			best = item
			bestRank = rank
		}
	}
	return best
}

func pickSportsCatalogLobby(items []TCGGameListItem) TCGGameListItem {
	best := items[0]
	bestRank := sportsCatalogLobbyRank(best)
	for _, item := range items[1:] {
		rank := sportsCatalogLobbyRank(item)
		if rank > bestRank {
			best = item
			bestRank = rank
			continue
		}
		if rank == bestRank && compareCatalogNames(item.GameName, best.GameName) < 0 {
			best = item
			bestRank = rank
		}
	}
	return best
}

func pickFishCatalogHero(items []TCGGameListItem) TCGGameListItem {
	best := items[0]
	bestRank := fishCatalogHeroRank(best)
	for _, item := range items[1:] {
		rank := fishCatalogHeroRank(item)
		if rank > bestRank {
			best = item
			bestRank = rank
			continue
		}
		if rank == bestRank && compareCatalogNames(item.GameName, best.GameName) < 0 {
			best = item
			bestRank = rank
		}
	}
	return best
}

func casinoCatalogHeroRank(item TCGGameListItem) int {
	score := 0
	if strings.TrimSpace(item.ShowIcon) != "" {
		score += 2
	}

	if containsCatalogKeyword(item, "lobby", "casino", "live", "room", "baccarat", "dragon", "roulette", "blackjack") {
		score += 5
	}

	if supportsDisplayPlatform(item.Platform) {
		score += 1
	}

	return score
}

func sportsCatalogLobbyRank(item TCGGameListItem) int {
	score := 0
	if containsCatalogKeyword(item, "football", "soccer", "bong da", "bongda") {
		score += 7
	}
	if containsCatalogKeyword(item, "sport", "sportsbook", "lobby", "saba", "ibc") {
		score += 4
	}
	if strings.TrimSpace(item.ShowIcon) != "" {
		score += 1
	}
	if supportsDisplayPlatform(item.Platform) {
		score += 1
	}
	return score
}

func fishCatalogHeroRank(item TCGGameListItem) int {
	score := 0
	if containsCatalogKeyword(item, "game_list", "game list", "lobby", "fish", "fishing", "ocean", "hunter") {
		score += 5
	}
	if strings.TrimSpace(item.ShowIcon) != "" {
		score += 2
	}
	if supportsDisplayPlatform(item.Platform) {
		score += 1
	}
	return score
}

func containsCatalogKeyword(item TCGGameListItem, keywords ...string) bool {
	values := []string{
		strings.ToLower(strings.TrimSpace(item.GameName)),
		strings.ToLower(strings.TrimSpace(item.TCGGameCode)),
		strings.ToLower(strings.TrimSpace(item.ProductCode)),
		strings.ToLower(strings.TrimSpace(item.GameSubType)),
	}

	for _, value := range values {
		for _, keyword := range keywords {
			if strings.Contains(value, strings.ToLower(strings.TrimSpace(keyword))) {
				return true
			}
		}
	}

	return false
}

func supportsDisplayPlatform(platform string) bool {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	if normalized == "" {
		return false
	}

	return strings.Contains(normalized, "html5") ||
		strings.Contains(normalized, "web") ||
		strings.Contains(normalized, "mobile") ||
		strings.Contains(normalized, "desktop")
}

func sortCatalogItems(items []TCGProviderCatalogCategoryItem) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left.ProductType != right.ProductType {
			return left.ProductType < right.ProductType
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		return compareCatalogNames(left.GameName, right.GameName) < 0
	})
}

func compareCatalogNames(left string, right string) int {
	trimmedLeft := strings.TrimSpace(left)
	trimmedRight := strings.TrimSpace(right)
	switch {
	case trimmedLeft < trimmedRight:
		return -1
	case trimmedLeft > trimmedRight:
		return 1
	default:
		return 0
	}
}

func parseTCGProductType(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

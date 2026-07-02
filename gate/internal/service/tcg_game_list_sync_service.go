package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gate/internal/integration/tcg"
	goredis "github.com/redis/go-redis/v9"
)

type TCGGameListSyncConfig struct {
	Enabled          bool
	Interval         time.Duration
	RedisKey         string
	ProductTypes     []int
	Platform         string
	ClientType       string
	GameTypes        []string
	GameTypeProducts map[string][]int
	Language         string
	Page             int
	PageSize         int
}

type TCGGameListSnapshot struct {
	Provider     string              `json:"provider"`
	Method       string              `json:"method"`
	Source       string              `json:"source"`
	ProductTypes []int               `json:"product_types"`
	GameTypes    []string            `json:"game_types"`
	ProductType  int                 `json:"product_type"`
	Platform     string              `json:"platform"`
	ClientType   string              `json:"client_type"`
	GameType     string              `json:"game_type"`
	Language     string              `json:"language,omitempty"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	Status       int                 `json:"status"`
	ErrorDesc    string              `json:"error_desc,omitempty"`
	SyncedAt     time.Time           `json:"synced_at"`
	Games        []TCGGameListItem   `json:"games"`
	PageInfo     TCGGameListPageInfo `json:"page_info"`
	Raw          map[string]any      `json:"raw,omitempty"`
}

type TCGGameListItem struct {
	DisplayStatus int    `json:"display_status"`
	GameType      string `json:"game_type"`
	GameName      string `json:"game_name"`
	TCGGameCode   string `json:"tcg_game_code"`
	ProductCode   string `json:"product_code"`
	ProductType   string `json:"product_type_value"`
	Platform      string `json:"platform"`
	GameSubType   string `json:"game_sub_type"`
	ShowIcon      string `json:"show_icon,omitempty"`
	TrialSupport  bool   `json:"trial_support"`
}

type TCGGameListPageInfo struct {
	TotalPage   int    `json:"total_page"`
	CurrentPage int    `json:"current_page"`
	TotalCount  string `json:"total_count"`
}

type TCGGameListSyncService struct {
	client *tcg.Client
	redis  *goredis.Client
	config TCGGameListSyncConfig
}

func NewTCGGameListSyncService(client *tcg.Client, redis *goredis.Client, config TCGGameListSyncConfig) *TCGGameListSyncService {
	return &TCGGameListSyncService{
		client: client,
		redis:  redis,
		config: config,
	}
}

func (s *TCGGameListSyncService) Start(ctx context.Context) {
	if s == nil {
		return
	}
	if !s.config.Enabled {
		log.Printf("[gate][tcg.game_list.sync.skip] reason=disabled")
		return
	}
	if s.client == nil {
		log.Printf("[gate][tcg.game_list.sync.skip] reason=client_nil")
		return
	}
	if s.redis == nil {
		log.Printf("[gate][tcg.game_list.sync.skip] reason=redis_nil")
		return
	}
	if strings.TrimSpace(s.config.RedisKey) == "" {
		log.Printf("[gate][tcg.game_list.sync.skip] reason=redis_key_empty")
		return
	}

	interval := s.config.Interval
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	go func() {
		s.syncOnce(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[gate][tcg.game_list.sync.stop] reason=context_done")
				return
			case <-ticker.C:
				s.syncOnce(ctx)
			}
		}
	}()
}

func (s *TCGGameListSyncService) syncOnce(ctx context.Context) {
	snapshot := s.fetchCombinedSnapshot(ctx)
	payload, err := json.Marshal(snapshot)
	if err != nil {
		log.Printf("[gate][tcg.game_list.sync.error] stage=marshal err=%v", err)
		return
	}

	if err := s.redis.Set(ctx, s.config.RedisKey, payload, intervalTTL(s.config.Interval)).Err(); err != nil {
		log.Printf("[gate][tcg.game_list.sync.error] stage=redis_set key=%s err=%v", s.config.RedisKey, err)
		return
	}

}

func (s *TCGGameListSyncService) fetchCombinedSnapshot(ctx context.Context) TCGGameListSnapshot {
	effectiveMap := s.effectiveGameTypeProducts()
	productTypes := aggregateProductTypes(effectiveMap)
	gameTypes := orderedGameTypes(effectiveMap)

	snapshot := TCGGameListSnapshot{
		Provider:     "tcg",
		Method:       "tgl",
		Source:       "tcg_api",
		ProductTypes: append([]int(nil), productTypes...),
		GameTypes:    append([]string(nil), gameTypes...),
		ProductType:  firstInt(productTypes),
		GameType:     firstString(gameTypes),
		Platform:     strings.TrimSpace(s.config.Platform),
		ClientType:   strings.TrimSpace(s.config.ClientType),
		Language:     strings.TrimSpace(s.config.Language),
		Page:         s.config.Page,
		PageSize:     s.config.PageSize,
		SyncedAt:     time.Now(),
		Games:        make([]TCGGameListItem, 0),
		PageInfo:     TCGGameListPageInfo{},
		Raw:          map[string]any{},
	}

	combinedTotalCount := 0
	combinedTotalPage := 0
	currentPage := 0
	successCount := 0
	seen := make(map[string]struct{})

	for _, gameType := range gameTypes {
		for _, productType := range effectiveMap[gameType] {
			result, err := s.client.ListGames(ctx, tcg.ListGamesRequest{
				ProductType: productType,
				Platform:    s.config.Platform,
				ClientType:  s.config.ClientType,
				GameType:    gameType,
				Language:    s.config.Language,
				Page:        s.config.Page,
				PageSize:    s.config.PageSize,
			})
			rawKey := strconv.Itoa(productType) + ":" + strings.ToUpper(strings.TrimSpace(gameType))
			if err != nil {
				log.Printf("[gate][tcg.game_list.sync.combo.error] product_type=%d game_type=%s err=%v", productType, gameType, err)
				snapshot.Raw[rawKey] = map[string]any{
					"status":     result.Status,
					"error_desc": strings.TrimSpace(result.ErrorDesc),
				}
				continue
			}

			successCount++
			partial := buildTCGGameListSnapshot(result, s.config, productType, gameType)
			snapshot.Raw[rawKey] = partial.Raw
			for _, item := range partial.Games {
				key := strings.TrimSpace(item.ProductType) + ":" + strings.TrimSpace(item.TCGGameCode)
				if key == ":" {
					key = strings.TrimSpace(item.GameType) + ":" + strings.TrimSpace(item.GameName)
				}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				snapshot.Games = append(snapshot.Games, item)
			}
			combinedTotalPage += partial.PageInfo.TotalPage
			currentPage = maxInt(currentPage, partial.PageInfo.CurrentPage)
			if parsedCount, err := strconv.Atoi(strings.TrimSpace(partial.PageInfo.TotalCount)); err == nil {
				combinedTotalCount += parsedCount
			}
		}
	}

	snapshot.Status = 0
	if successCount == 0 {
		snapshot.Status = 1
		snapshot.ErrorDesc = "all product type sync requests failed"
	}
	if combinedTotalCount > 0 {
		snapshot.PageInfo.TotalCount = strconv.Itoa(combinedTotalCount)
	}
	snapshot.PageInfo.TotalPage = combinedTotalPage
	snapshot.PageInfo.CurrentPage = currentPage

	if len(snapshot.Games) == 0 {
		snapshot.Games = buildStaticTCGGameCatalog(effectiveMap)
		if len(snapshot.Games) > 0 {
			snapshot.Source = "static_docs_fallback"
			snapshot.PageInfo = TCGGameListPageInfo{
				TotalPage:   1,
				CurrentPage: 1,
				TotalCount:  strconv.Itoa(len(snapshot.Games)),
			}
		}
	}

	return snapshot
}

func (s *TCGGameListSyncService) effectiveGameTypeProducts() map[string][]int {
	if len(s.config.GameTypeProducts) > 0 {
		result := make(map[string][]int, len(s.config.GameTypeProducts))
		for _, gameType := range orderedGameTypes(s.config.GameTypeProducts) {
			productTypes := dedupeIntList(s.config.GameTypeProducts[gameType])
			if len(productTypes) == 0 {
				continue
			}
			result[gameType] = productTypes
		}
		if len(result) > 0 {
			return result
		}
	}

	productTypes := dedupeIntList(s.config.ProductTypes)
	if len(productTypes) == 0 {
		productTypes = []int{7}
	}

	gameTypes := dedupeStringList(s.config.GameTypes)
	if len(gameTypes) == 0 {
		gameTypes = []string{"RNG"}
	}

	result := make(map[string][]int, len(gameTypes))
	for _, gameType := range gameTypes {
		result[gameType] = append([]int(nil), productTypes...)
	}
	return result
}

func buildTCGGameListSnapshot(result tcg.BusinessResponse, config TCGGameListSyncConfig, productType int, gameType string) TCGGameListSnapshot {
	snapshot := TCGGameListSnapshot{
		Provider:     "tcg",
		Method:       "tgl",
		Source:       "tcg_api",
		ProductTypes: []int{productType},
		GameTypes:    []string{gameType},
		ProductType:  productType,
		Platform:     strings.TrimSpace(config.Platform),
		ClientType:   strings.TrimSpace(config.ClientType),
		GameType:     strings.TrimSpace(gameType),
		Language:     strings.TrimSpace(config.Language),
		Page:         config.Page,
		PageSize:     config.PageSize,
		Status:       result.Status,
		ErrorDesc:    strings.TrimSpace(result.ErrorDesc),
		SyncedAt:     time.Now(),
		PageInfo:     extractPageInfo(result.RawPayload),
		Raw:          result.RawPayload,
	}

	rawGames, _ := result.RawPayload["games"].([]any)
	snapshot.Games = make([]TCGGameListItem, 0, len(rawGames))
	for _, rawGame := range rawGames {
		gameMap, ok := rawGame.(map[string]any)
		if !ok {
			continue
		}

		snapshot.Games = append(snapshot.Games, TCGGameListItem{
			DisplayStatus: extractInt(gameMap, "displayStatus"),
			GameType:      firstNonEmptyStringFromMap(gameMap, "gameType"),
			GameName:      firstNonEmptyStringFromMap(gameMap, "gameName"),
			TCGGameCode:   firstNonEmptyStringFromMap(gameMap, "tcgGameCode"),
			ProductCode:   firstNonEmptyStringFromMap(gameMap, "productCode"),
			ProductType:   firstNonEmptyStringFromMap(gameMap, "productType"),
			Platform:      firstNonEmptyStringFromMap(gameMap, "platform"),
			GameSubType:   firstNonEmptyStringFromMap(gameMap, "gameSubType"),
			ShowIcon:      firstNonEmptyStringFromMap(gameMap, "showIcon"),
			TrialSupport:  extractBool(gameMap, "trialSupport"),
		})
	}

	return snapshot
}

func firstInt(items []int) int {
	if len(items) == 0 {
		return 0
	}
	return items[0]
}

func firstString(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func orderedGameTypes(items map[string][]int) []string {
	knownOrder := []string{"RNG", "FISH", "LIVE", "PVP", "SPORTS", "ELOTT"}
	ordered := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, candidate := range knownOrder {
		if _, ok := items[candidate]; ok {
			ordered = append(ordered, candidate)
			seen[candidate] = struct{}{}
		}
	}
	for candidate := range items {
		if _, ok := seen[candidate]; ok {
			continue
		}
		ordered = append(ordered, candidate)
	}
	return ordered
}

func aggregateProductTypes(items map[string][]int) []int {
	seen := map[int]struct{}{}
	ordered := make([]int, 0)
	for _, gameType := range orderedGameTypes(items) {
		for _, productType := range items[gameType] {
			if _, ok := seen[productType]; ok {
				continue
			}
			seen[productType] = struct{}{}
			ordered = append(ordered, productType)
		}
	}
	return ordered
}

func dedupeIntList(items []int) []int {
	seen := map[int]struct{}{}
	result := make([]int, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func dedupeStringList(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.ToUpper(strings.TrimSpace(item))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func extractPageInfo(payload map[string]any) TCGGameListPageInfo {
	pageInfoMap, _ := payload["page_info"].(map[string]any)
	if pageInfoMap == nil {
		return TCGGameListPageInfo{}
	}

	return TCGGameListPageInfo{
		TotalPage:   extractInt(pageInfoMap, "totalPage"),
		CurrentPage: extractInt(pageInfoMap, "currentPage"),
		TotalCount:  firstNonEmptyStringFromMap(pageInfoMap, "totalCount"),
	}
}

func extractInt(payload map[string]any, key string) int {
	value, ok := payload[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}

	return 0
}

func extractBool(payload map[string]any, key string) bool {
	value, ok := payload[key]
	if !ok || value == nil {
		return false
	}

	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func firstNonEmptyStringFromMap(payload map[string]any, key string) string {
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}

	trimmed := strings.TrimSpace(fmt.Sprint(value))
	if trimmed == "<nil>" {
		return ""
	}
	return trimmed
}

func intervalTTL(interval time.Duration) time.Duration {
	if interval <= 0 {
		return 15 * time.Minute
	}

	return interval * 3
}

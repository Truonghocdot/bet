package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	goredis "github.com/redis/go-redis/v9"
)

type ProviderGameCatalogConfig struct {
	RedisKey string
}

type ProviderGameCatalogFilter struct {
	Category        string
	ProductType     int
	IncludeChildren bool
}

type ProviderGameCatalogService struct {
	redis  *goredis.Client
	config ProviderGameCatalogConfig
}

type providerGameCatalogSnapshot struct {
	Provider     string                           `json:"provider"`
	Method       string                           `json:"method"`
	Source       string                           `json:"source"`
	ProductTypes []int                            `json:"product_types"`
	GameTypes    []string                         `json:"game_types"`
	ProductType  int                              `json:"product_type"`
	Platform     string                           `json:"platform"`
	ClientType   string                           `json:"client_type"`
	GameType     string                           `json:"game_type"`
	Language     string                           `json:"language,omitempty"`
	SyncedAt     time.Time                        `json:"synced_at"`
	Categories   []providerGameCatalogCategory    `json:"categories"`
	PageInfo     game.ProviderGameCatalogPageInfo `json:"page_info"`
}

type providerGameCatalogCategory struct {
	Key   string                            `json:"key"`
	Label string                            `json:"label"`
	Items []providerGameCatalogCategoryItem `json:"items"`
}

type providerGameCatalogCategoryItem struct {
	Kind             string                         `json:"kind"`
	DisplayStatus    int                            `json:"display_status"`
	GameType         string                         `json:"game_type"`
	GameName         string                         `json:"game_name"`
	TCGGameCode      string                         `json:"tcg_game_code"`
	ProductCode      string                         `json:"product_code"`
	ProductType      int                            `json:"product_type"`
	ProductTypeValue string                         `json:"product_type_value"`
	Platform         string                         `json:"platform"`
	GameSubType      string                         `json:"game_sub_type"`
	ShowIcon         string                         `json:"show_icon,omitempty"`
	TrialSupport     bool                           `json:"trial_support"`
	ChildCount       int                            `json:"child_count,omitempty"`
	Children         []game.ProviderGameCatalogItem `json:"children,omitempty"`
}

func NewProviderGameCatalogService(redis *goredis.Client, config ProviderGameCatalogConfig) *ProviderGameCatalogService {
	return &ProviderGameCatalogService{
		redis:  redis,
		config: config,
	}
}

func (s *ProviderGameCatalogService) GetTCGGameCatalog(ctx context.Context, filter ProviderGameCatalogFilter) (game.ProviderGameCatalogResponse, error) {
	if s.redis == nil {
		return game.ProviderGameCatalogResponse{}, fmt.Errorf("redis disabled")
	}
	if strings.TrimSpace(s.config.RedisKey) == "" {
		return game.ProviderGameCatalogResponse{}, fmt.Errorf("game catalog redis key is required")
	}

	raw, err := s.redis.Get(ctx, s.config.RedisKey).Result()
	if err != nil {
		if err == goredis.Nil {
			return game.ProviderGameCatalogResponse{}, fmt.Errorf("provider game catalog not ready")
		}
		return game.ProviderGameCatalogResponse{}, err
	}

	var snapshot providerGameCatalogSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return game.ProviderGameCatalogResponse{}, err
	}

	categories := filterProviderCatalogCategories(snapshot.Categories, filter)

	return game.ProviderGameCatalogResponse{
		Message:      "Lấy danh sách game nhà cung cấp thành công",
		Provider:     strings.TrimSpace(snapshot.Provider),
		Method:       strings.TrimSpace(snapshot.Method),
		Source:       strings.TrimSpace(snapshot.Source),
		ProductTypes: append([]int(nil), snapshot.ProductTypes...),
		GameTypes:    append([]string(nil), snapshot.GameTypes...),
		ProductType:  snapshot.ProductType,
		Platform:     strings.TrimSpace(snapshot.Platform),
		ClientType:   strings.TrimSpace(snapshot.ClientType),
		GameType:     strings.TrimSpace(snapshot.GameType),
		Language:     strings.TrimSpace(snapshot.Language),
		SyncedAt:     snapshot.SyncedAt,
		Categories:   categories,
		PageInfo:     snapshot.PageInfo,
	}, nil
}

func filterProviderCatalogCategories(
	categories []providerGameCatalogCategory,
	filter ProviderGameCatalogFilter,
) []game.ProviderGameCatalogCategory {
	selectedCategory := strings.ToLower(strings.TrimSpace(filter.Category))
	result := make([]game.ProviderGameCatalogCategory, 0, len(categories))

	for _, category := range categories {
		categoryKey := strings.ToLower(strings.TrimSpace(category.Key))
		if selectedCategory != "" && categoryKey != selectedCategory {
			continue
		}

		items := make([]game.ProviderGameCatalogCategoryItem, 0, len(category.Items))
		for _, item := range category.Items {
			if filter.ProductType > 0 && item.ProductType != filter.ProductType && parseProviderProductTypeValue(item.ProductTypeValue) != filter.ProductType {
				continue
			}

			children := make([]game.ProviderGameCatalogItem, 0, len(item.Children))
			if filter.IncludeChildren {
				children = append(children, item.Children...)
			}

			items = append(items, game.ProviderGameCatalogCategoryItem{
				Kind:             strings.TrimSpace(item.Kind),
				DisplayStatus:    item.DisplayStatus,
				GameType:         strings.TrimSpace(item.GameType),
				GameName:         strings.TrimSpace(item.GameName),
				TCGGameCode:      strings.TrimSpace(item.TCGGameCode),
				ProductCode:      strings.TrimSpace(item.ProductCode),
				ProductType:      item.ProductType,
				ProductTypeValue: strings.TrimSpace(item.ProductTypeValue),
				Platform:         strings.TrimSpace(item.Platform),
				GameSubType:      strings.TrimSpace(item.GameSubType),
				ShowIcon:         strings.TrimSpace(item.ShowIcon),
				TrialSupport:     item.TrialSupport,
				ChildCount:       item.ChildCount,
				Children:         children,
			})
		}

		if len(items) == 0 {
			continue
		}

		result = append(result, game.ProviderGameCatalogCategory{
			Key:   strings.TrimSpace(category.Key),
			Label: strings.TrimSpace(category.Label),
			Items: items,
		})
	}

	return result
}

func parseProviderProductTypeValue(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gin/internal/domain/game"
	goredis "github.com/redis/go-redis/v9"
)

type ProviderGameCatalogConfig struct {
	RedisKey string
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
	Games        []game.ProviderGameCatalogItem   `json:"games"`
	PageInfo     game.ProviderGameCatalogPageInfo `json:"page_info"`
}

func NewProviderGameCatalogService(redis *goredis.Client, config ProviderGameCatalogConfig) *ProviderGameCatalogService {
	return &ProviderGameCatalogService{
		redis:  redis,
		config: config,
	}
}

func (s *ProviderGameCatalogService) GetTCGGameCatalog(ctx context.Context) (game.ProviderGameCatalogResponse, error) {
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
		Items:        snapshot.Games,
		PageInfo:     snapshot.PageInfo,
	}, nil
}

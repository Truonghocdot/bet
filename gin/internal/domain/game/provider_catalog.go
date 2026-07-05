package game

import "time"

type ProviderGameCatalogResponse struct {
	Message      string                        `json:"message"`
	Provider     string                        `json:"provider"`
	Method       string                        `json:"method"`
	Source       string                        `json:"source"`
	ProductTypes []int                         `json:"product_types"`
	GameTypes    []string                      `json:"game_types"`
	ProductType  int                           `json:"product_type"`
	Platform     string                        `json:"platform"`
	ClientType   string                        `json:"client_type"`
	GameType     string                        `json:"game_type"`
	Language     string                        `json:"language,omitempty"`
	SyncedAt     time.Time                     `json:"synced_at"`
	Categories   []ProviderGameCatalogCategory `json:"categories"`
	PageInfo     ProviderGameCatalogPageInfo   `json:"page_info"`
}

type ProviderGameCatalogCategory struct {
	Key   string                            `json:"key"`
	Label string                            `json:"label"`
	Items []ProviderGameCatalogCategoryItem `json:"items"`
}

type ProviderGameCatalogCategoryItem struct {
	Kind             string                    `json:"kind"`
	DisplayStatus    int                       `json:"display_status"`
	GameType         string                    `json:"game_type"`
	GameName         string                    `json:"game_name"`
	TCGGameCode      string                    `json:"tcg_game_code"`
	ProductCode      string                    `json:"product_code"`
	ProductType      int                       `json:"product_type"`
	ProductTypeValue string                    `json:"product_type_value"`
	Platform         string                    `json:"platform"`
	GameSubType      string                    `json:"game_sub_type"`
	ShowIcon         string                    `json:"show_icon,omitempty"`
	TrialSupport     bool                      `json:"trial_support"`
	ChildCount       int                       `json:"child_count,omitempty"`
	Children         []ProviderGameCatalogItem `json:"children,omitempty"`
}

type ProviderGameCatalogItem struct {
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

type ProviderGameCatalogPageInfo struct {
	TotalPage   int    `json:"total_page"`
	CurrentPage int    `json:"current_page"`
	TotalCount  string `json:"total_count"`
}

type ProviderGameLaunchRequest struct {
	ProductType int    `json:"product_type"`
	GameType    string `json:"game_type"`
	GameCode    string `json:"game_code"`
	Name        string `json:"name"`
}

type ProviderGameLaunchResponse struct {
	Message           string `json:"message"`
	GameURL           string `json:"game_url"`
	ProductType       int    `json:"product_type"`
	GameType          string `json:"game_type"`
	TransferredAmount string `json:"transferred_amount"`
	Source            string `json:"source"`
}

type ProviderGameCloseActiveResponse struct {
	Message     string `json:"message"`
	ProductType int    `json:"product_type"`
	GameType    string `json:"game_type"`
	Swept       bool   `json:"swept"`
	ReferenceNo string `json:"reference_no,omitempty"`
}

type ProviderGameLaunchMeta struct {
	IP        string
	UserAgent string
}

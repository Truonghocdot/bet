package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type RegisterTCGPlayerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Currency string `json:"currency,omitempty"`
}

type BalanceTCGRequest struct {
	Username    string `json:"username"`
	ProductType int    `json:"product_type"`
}

type BalanceTCGResponse struct {
	Status    int            `json:"status"`
	ErrorDesc string         `json:"error_desc,omitempty"`
	Balance   string         `json:"balance,omitempty"`
	Raw       map[string]any `json:"raw,omitempty"`
}

type TransferTCGRequest struct {
	Username    string `json:"username"`
	ProductType int    `json:"product_type"`
	FundType    string `json:"fund_type"`
	Amount      string `json:"amount"`
	ReferenceNo string `json:"reference_no"`
}

type TransferTCGResponse struct {
	Status      int            `json:"status"`
	ErrorDesc   string         `json:"error_desc,omitempty"`
	ReferenceNo string         `json:"reference_no,omitempty"`
	Raw         map[string]any `json:"raw,omitempty"`
}

type TransferStatusTCGRequest struct {
	ProductType int    `json:"product_type"`
	ReferenceNo string `json:"reference_no"`
}

type TransferStatusTCGResponse struct {
	Status            int            `json:"status"`
	ErrorDesc         string         `json:"error_desc,omitempty"`
	TransactionStatus string         `json:"transaction_status,omitempty"`
	TransactionDetail map[string]any `json:"transaction_details,omitempty"`
	Raw               map[string]any `json:"raw,omitempty"`
}

type LaunchGameTCGRequest struct {
	Username       string         `json:"username"`
	ProductType    int            `json:"product_type"`
	IPAddress      string         `json:"ip_address"`
	Platform       string         `json:"platform"`
	GameMode       string         `json:"game_mode"`
	GameCode       string         `json:"game_code"`
	Language       string         `json:"language,omitempty"`
	Nickname       string         `json:"nickname,omitempty"`
	VIPLevel       *int           `json:"vip_level,omitempty"`
	View           string         `json:"view,omitempty"`
	BackURL        string         `json:"back_url,omitempty"`
	LotteryBetMode string         `json:"lottery_bet_mode,omitempty"`
	Series         map[string]any `json:"series,omitempty"`
}

type LaunchGameTCGResponse struct {
	Status     int            `json:"status"`
	ErrorDesc  string         `json:"error_desc,omitempty"`
	GameURL    string         `json:"game_url,omitempty"`
	PTUsername string         `json:"pt_username,omitempty"`
	PTPassword string         `json:"pt_password,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}

type TCGClient struct {
	baseURL       string
	internalToken string
	client        *http.Client
}

func NewTCGClient(baseURL, internalToken string) *TCGClient {
	return &TCGClient{
		baseURL:       strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		internalToken: strings.TrimSpace(internalToken),
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *TCGClient) RegisterPlayer(ctx context.Context, request RegisterTCGPlayerRequest) error {
	_, err := postTCGJSON[map[string]string](c, ctx, "/internal/v1/tcg/players/register", request)
	return err
}

func (c *TCGClient) GetBalance(ctx context.Context, request BalanceTCGRequest) (BalanceTCGResponse, error) {
	return postTCGJSON[BalanceTCGResponse](c, ctx, "/internal/v1/tcg/wallets/balance", request)
}

func (c *TCGClient) Transfer(ctx context.Context, request TransferTCGRequest) (TransferTCGResponse, error) {
	return postTCGJSON[TransferTCGResponse](c, ctx, "/internal/v1/tcg/wallets/transfer", request)
}

func (c *TCGClient) TransferOutAll(ctx context.Context, request TransferTCGRequest) (TransferTCGResponse, error) {
	return postTCGJSON[TransferTCGResponse](c, ctx, "/internal/v1/tcg/wallets/transfer-out-all", request)
}

func (c *TCGClient) CheckTransferStatus(ctx context.Context, request TransferStatusTCGRequest) (TransferStatusTCGResponse, error) {
	return postTCGJSON[TransferStatusTCGResponse](c, ctx, "/internal/v1/tcg/wallets/transfer-status", request)
}

func (c *TCGClient) LaunchGame(ctx context.Context, request LaunchGameTCGRequest) (LaunchGameTCGResponse, error) {
	return postTCGJSON[LaunchGameTCGResponse](c, ctx, "/internal/v1/tcg/games/launch", request)
}

func postTCGJSON[T any](c *TCGClient, ctx context.Context, path string, request any) (T, error) {
	var zero T
	if c.baseURL == "" {
		log.Printf("[gin][gate.tcg.error] action=post_json reason=base_url_missing path=%s", path)
		return zero, fmt.Errorf("gate base url is required")
	}

	body, err := json.Marshal(request)
	if err != nil {
		log.Printf("[gin][gate.tcg.error] action=post_json reason=marshal_failed path=%s err=%v", path, err)
		return zero, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[gin][gate.tcg.error] action=post_json reason=new_request_failed path=%s err=%v", path, err)
		return zero, err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Internal-Token", c.internalToken)

	log.Printf("[gin][gate.tcg.request] path=%s body=%s", path, strings.TrimSpace(string(body)))

	response, err := c.client.Do(httpRequest)
	if err != nil {
		log.Printf("[gin][gate.tcg.error] action=post_json reason=do_failed path=%s err=%v", path, err)
		return zero, err
	}
	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	bodyText := strings.TrimSpace(string(bodyBytes))
	log.Printf("[gin][gate.tcg.response] path=%s status=%d body=%s", path, response.StatusCode, bodyText)

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return zero, fmt.Errorf("gate tcg request returned status %d: %s", response.StatusCode, bodyText)
	}

	if len(bodyBytes) == 0 {
		return zero, nil
	}

	var parsed T
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return zero, err
	}

	return parsed, nil
}

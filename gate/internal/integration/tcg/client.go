package tcg

import (
	"bytes"
	"context"
	"crypto/des"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	merchantCode string
	desKey       string
	signKey      string
	client       *http.Client
}

type CreatePlayerRequest struct {
	Username string
	Password string
	Currency string
}

type GetBalanceRequest struct {
	Username    string
	ProductType int
}

type FundTransferRequest struct {
	Username    string
	ProductType int
	FundType    string
	Amount      string
	ReferenceNo string
}

type TransferOutAllRequest struct {
	Username    string
	ProductType int
	ReferenceNo string
}

type CheckTransferStatusRequest struct {
	ProductType int
	ReferenceNo string
}

type ListGamesRequest struct {
	ProductType int
	Platform    string
	ClientType  string
	GameType    string
	Language    string
	Page        int
	PageSize    int
}

type LaunchGameRequest struct {
	Username       string
	ProductType    int
	IPAddress      string
	Platform       string
	GameMode       string
	GameCode       string
	Language       string
	Nickname       string
	VIPLevel       *int
	View           string
	BackURL        string
	LotteryBetMode string
	Series         any
}

type BusinessResponse struct {
	Status     int
	ErrorDesc  string
	RawPayload map[string]any
}

func NewClient(baseURL, merchantCode, desKey, signKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	normalizedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	normalizedBaseURL = strings.TrimSuffix(normalizedBaseURL, "/doBusiness.do")

	return &Client{
		baseURL:      normalizedBaseURL,
		merchantCode: strings.TrimSpace(merchantCode),
		desKey:       strings.TrimSpace(desKey),
		signKey:      strings.TrimSpace(signKey),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) CreatePlayer(ctx context.Context, request CreatePlayerRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":   "cm",
		"username": strings.TrimSpace(request.Username),
		"password": strings.TrimSpace(request.Password),
	}
	if currency := strings.TrimSpace(request.Currency); currency != "" {
		params["currency"] = currency
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) GetBalance(ctx context.Context, request GetBalanceRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "gb",
		"username":     strings.TrimSpace(request.Username),
		"product_type": request.ProductType,
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) FundTransfer(ctx context.Context, request FundTransferRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "ft",
		"username":     strings.TrimSpace(request.Username),
		"product_type": request.ProductType,
		"fund_type":    strings.TrimSpace(request.FundType),
		"amount":       strings.TrimSpace(request.Amount),
		"reference_no": strings.TrimSpace(request.ReferenceNo),
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) TransferOutAll(ctx context.Context, request TransferOutAllRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "ftoa",
		"username":     strings.TrimSpace(request.Username),
		"product_type": request.ProductType,
		"reference_no": strings.TrimSpace(request.ReferenceNo),
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) CheckTransferStatus(ctx context.Context, request CheckTransferStatusRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "cs",
		"product_type": request.ProductType,
		"ref_no":       strings.TrimSpace(request.ReferenceNo),
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) ListGames(ctx context.Context, request ListGamesRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "tgl",
		"product_type": request.ProductType,
		"platform":     strings.TrimSpace(request.Platform),
		"client_type":  strings.TrimSpace(request.ClientType),
		"game_type":    strings.TrimSpace(request.GameType),
		"page":         request.Page,
		"page_size":    request.PageSize,
	}
	if language := strings.TrimSpace(request.Language); language != "" {
		params["language"] = language
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) LaunchGame(ctx context.Context, request LaunchGameRequest) (BusinessResponse, error) {
	params := map[string]any{
		"method":       "lg",
		"username":     strings.TrimSpace(request.Username),
		"product_type": request.ProductType,
		"ip_address":   strings.TrimSpace(request.IPAddress),
		"platform":     strings.TrimSpace(request.Platform),
		"game_mode":    strings.TrimSpace(request.GameMode),
		"game_code":    strings.TrimSpace(request.GameCode),
	}
	if language := strings.TrimSpace(request.Language); language != "" {
		params["language"] = language
	}
	if nickname := strings.TrimSpace(request.Nickname); nickname != "" {
		params["nickname"] = nickname
	}
	if request.VIPLevel != nil {
		params["vip_level"] = *request.VIPLevel
	}
	if view := strings.TrimSpace(request.View); view != "" {
		params["view"] = view
	}
	if backURL := strings.TrimSpace(request.BackURL); backURL != "" {
		params["back_url"] = backURL
	}
	if lotteryBetMode := strings.TrimSpace(request.LotteryBetMode); lotteryBetMode != "" {
		params["lottery_bet_mode"] = lotteryBetMode
	}
	if request.Series != nil {
		params["series"] = request.Series
	}

	return c.doBusiness(ctx, params)
}

func (c *Client) doBusiness(ctx context.Context, params map[string]any) (BusinessResponse, error) {
	method := strings.TrimSpace(fmt.Sprint(params["method"]))
	shouldLogInfo := method != "tgl"

	if c.baseURL == "" {
		log.Printf("[gate][tcg.config.error] method=%s reason=base_url_missing", method)
		return BusinessResponse{}, fmt.Errorf("tcg base url is required")
	}
	if c.merchantCode == "" {
		log.Printf("[gate][tcg.config.error] method=%s reason=merchant_code_missing", method)
		return BusinessResponse{}, fmt.Errorf("tcg merchant code is required")
	}
	if c.desKey == "" {
		log.Printf("[gate][tcg.config.error] method=%s reason=des_key_missing", method)
		return BusinessResponse{}, fmt.Errorf("tcg merchant des key is required")
	}
	if c.signKey == "" {
		log.Printf("[gate][tcg.config.error] method=%s reason=sign_key_missing", method)
		return BusinessResponse{}, fmt.Errorf("tcg merchant sign key is required")
	}

	rawParams, err := json.Marshal(params)
	if err != nil {
		log.Printf("[gate][tcg.encode.error] method=%s stage=marshal err=%v", method, err)
		return BusinessResponse{}, err
	}

	encryptedParams, err := encryptDESPKCS5Base64(rawParams, c.desKey)
	if err != nil {
		log.Printf("[gate][tcg.encode.error] method=%s stage=encrypt err=%v", method, err)
		return BusinessResponse{}, err
	}

	signature := sha256Hex(encryptedParams + c.signKey)

	form := url.Values{}
	form.Set("merchant_code", c.merchantCode)
	form.Set("params", encryptedParams)
	form.Set("sign", signature)

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/doBusiness.do",
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		log.Printf("[gate][tcg.request.error] method=%s stage=new_request err=%v", method, err)
		return BusinessResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if shouldLogInfo {
		log.Printf(
			"[gate][tcg.request] method=%s url=%s merchant=%s payload=%s encrypted_len=%d sign_prefix=%s",
			method,
			c.baseURL+"/doBusiness.do",
			maskValue(c.merchantCode, 2, 2),
			string(rawParams),
			len(encryptedParams),
			maskValue(signature, 6, 4),
		)
	}

	response, err := c.client.Do(httpRequest)
	if err != nil {
		log.Printf("[gate][tcg.request.error] method=%s stage=do err=%v", method, err)
		return BusinessResponse{}, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("[gate][tcg.response.error] method=%s stage=read_body err=%v", method, err)
		return BusinessResponse{}, err
	}

	bodyText := strings.TrimSpace(string(bodyBytes))
	if shouldLogInfo {
		log.Printf(
			"[gate][tcg.response] method=%s http_status=%d body=%s",
			method,
			response.StatusCode,
			clipString(bodyText, 1200),
		)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf("[gate][tcg.response.error] method=%s stage=http_status status=%d", method, response.StatusCode)
		return BusinessResponse{}, fmt.Errorf("tcg returned status %d body=%s", response.StatusCode, bodyText)
	}

	var payload map[string]any
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("[gate][tcg.response.error] method=%s stage=unmarshal err=%v", method, err)
		return BusinessResponse{}, err
	}

	result := BusinessResponse{
		Status:     extractStatus(payload),
		ErrorDesc:  firstNonEmptyString(payload, "error_desc", "error_message", "message"),
		RawPayload: payload,
	}

	if result.Status != 0 {
		if result.ErrorDesc == "" {
			result.ErrorDesc = "tcg business error"
		}
		log.Printf("[gate][tcg.business.error] method=%s status=%d desc=%s", method, result.Status, result.ErrorDesc)
		return result, fmt.Errorf("tcg business error status=%d desc=%s", result.Status, result.ErrorDesc)
	}

	if shouldLogInfo {
		log.Printf("[gate][tcg.business.ok] method=%s status=%d", method, result.Status)
	}

	return result, nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum[:])
}

func encryptDESPKCS5Base64(plain []byte, key string) (string, error) {
	keyBytes := []byte(strings.TrimSpace(key))
	if len(keyBytes) < 8 {
		return "", fmt.Errorf("tcg des key must be at least 8 bytes")
	}

	block, err := des.NewCipher(keyBytes[:8])
	if err != nil {
		return "", err
	}

	padded := pkcs5Pad(plain, block.BlockSize())
	encrypted := make([]byte, len(padded))
	for offset := 0; offset < len(padded); offset += block.BlockSize() {
		block.Encrypt(encrypted[offset:offset+block.BlockSize()], padded[offset:offset+block.BlockSize()])
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func pkcs5Pad(value []byte, blockSize int) []byte {
	padding := blockSize - (len(value) % blockSize)
	if padding == 0 {
		padding = blockSize
	}

	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(value, padText...)
}

func extractStatus(payload map[string]any) int {
	if payload == nil {
		return -1
	}

	value, ok := payload["status"]
	if !ok {
		return -1
	}

	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "0" {
			return 0
		}
	}

	return -1
}

func firstNonEmptyString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok || value == nil {
			continue
		}

		trimmed := strings.TrimSpace(fmt.Sprint(value))
		if trimmed != "" && trimmed != "<nil>" {
			return trimmed
		}
	}

	return ""
}

func clipString(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 || len(trimmed) <= max {
		return trimmed
	}

	return trimmed[:max] + "...(truncated)"
}

func maskValue(value string, prefix int, suffix int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if prefix < 0 {
		prefix = 0
	}
	if suffix < 0 {
		suffix = 0
	}

	if len(trimmed) <= prefix+suffix {
		return trimmed
	}

	return trimmed[:prefix] + "..." + trimmed[len(trimmed)-suffix:]
}

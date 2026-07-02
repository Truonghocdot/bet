package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gate/internal/integration/tcg"
)

type TCGConfig struct {
	Enabled bool
}

type RegisterTCGPlayerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Currency string `json:"currency,omitempty"`
}

type TCGBalanceRequest struct {
	Username    string `json:"username"`
	ProductType int    `json:"product_type"`
}

type TCGBalanceResponse struct {
	Status    int            `json:"status"`
	ErrorDesc string         `json:"error_desc,omitempty"`
	Balance   string         `json:"balance,omitempty"`
	Raw       map[string]any `json:"raw,omitempty"`
}

type TCGTransferRequest struct {
	Username    string `json:"username"`
	ProductType int    `json:"product_type"`
	FundType    string `json:"fund_type"`
	Amount      string `json:"amount"`
	ReferenceNo string `json:"reference_no"`
}

type TCGTransferResponse struct {
	Status      int            `json:"status"`
	ErrorDesc   string         `json:"error_desc,omitempty"`
	ReferenceNo string         `json:"reference_no,omitempty"`
	Raw         map[string]any `json:"raw,omitempty"`
}

type TCGTransferStatusRequest struct {
	ProductType int    `json:"product_type"`
	ReferenceNo string `json:"reference_no"`
}

type TCGTransferStatusResponse struct {
	Status            int            `json:"status"`
	ErrorDesc         string         `json:"error_desc,omitempty"`
	TransactionStatus string         `json:"transaction_status,omitempty"`
	TransactionDetail map[string]any `json:"transaction_details,omitempty"`
	Raw               map[string]any `json:"raw,omitempty"`
}

type TCGLaunchRequest struct {
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

type TCGLaunchResponse struct {
	Status     int            `json:"status"`
	ErrorDesc  string         `json:"error_desc,omitempty"`
	GameURL    string         `json:"game_url,omitempty"`
	PTUsername string         `json:"pt_username,omitempty"`
	PTPassword string         `json:"pt_password,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
}

type TCGService struct {
	client  *tcg.Client
	enabled bool
}

func NewTCGService(client *tcg.Client, config TCGConfig) *TCGService {
	return &TCGService{
		client:  client,
		enabled: config.Enabled,
	}
}

func (s *TCGService) RegisterPlayer(ctx context.Context, request RegisterTCGPlayerRequest) error {
	if !s.enabled {
		log.Printf("[gate][tcg.service.error] action=register_player reason=integration_disabled username=%s", strings.TrimSpace(request.Username))
		return fmt.Errorf("tcg integration is disabled")
	}
	if s.client == nil {
		log.Printf("[gate][tcg.service.error] action=register_player reason=client_missing username=%s", strings.TrimSpace(request.Username))
		return fmt.Errorf("tcg client is not configured")
	}

	username := strings.TrimSpace(request.Username)
	password := strings.TrimSpace(request.Password)
	if username == "" || password == "" {
		log.Printf("[gate][tcg.service.error] action=register_player reason=missing_required_fields username=%s", username)
		return fmt.Errorf("username and password are required")
	}

	log.Printf(
		"[gate][tcg.service.start] action=register_player username=%s currency=%s password_len=%d",
		username,
		strings.TrimSpace(request.Currency),
		len(password),
	)

	_, err := s.client.CreatePlayer(ctx, tcg.CreatePlayerRequest{
		Username: username,
		Password: password,
		Currency: strings.TrimSpace(request.Currency),
	})
	if err != nil {
		log.Printf("[gate][tcg.service.error] action=register_player username=%s err=%v", username, err)
		return err
	}

	log.Printf("[gate][tcg.service.ok] action=register_player username=%s", username)
	return err
}

func (s *TCGService) GetBalance(ctx context.Context, request TCGBalanceRequest) (TCGBalanceResponse, error) {
	if err := s.ensureReady("get_balance", request.Username); err != nil {
		return TCGBalanceResponse{}, err
	}
	result, err := s.client.GetBalance(ctx, tcg.GetBalanceRequest{
		Username:    strings.TrimSpace(request.Username),
		ProductType: request.ProductType,
	})
	response := TCGBalanceResponse{
		Status:    result.Status,
		ErrorDesc: strings.TrimSpace(result.ErrorDesc),
		Balance:   firstNonEmptyTCGValue(result.RawPayload, "balance"),
		Raw:       result.RawPayload,
	}
	return response, err
}

func (s *TCGService) Transfer(ctx context.Context, request TCGTransferRequest) (TCGTransferResponse, error) {
	if err := s.ensureReady("transfer", request.Username); err != nil {
		return TCGTransferResponse{}, err
	}
	result, err := s.client.FundTransfer(ctx, tcg.FundTransferRequest{
		Username:    strings.TrimSpace(request.Username),
		ProductType: request.ProductType,
		FundType:    strings.TrimSpace(request.FundType),
		Amount:      strings.TrimSpace(request.Amount),
		ReferenceNo: strings.TrimSpace(request.ReferenceNo),
	})
	response := TCGTransferResponse{
		Status:      result.Status,
		ErrorDesc:   strings.TrimSpace(result.ErrorDesc),
		ReferenceNo: strings.TrimSpace(request.ReferenceNo),
		Raw:         result.RawPayload,
	}
	return response, err
}

func (s *TCGService) TransferOutAll(ctx context.Context, request TCGTransferRequest) (TCGTransferResponse, error) {
	if err := s.ensureReady("transfer_out_all", request.Username); err != nil {
		return TCGTransferResponse{}, err
	}
	result, err := s.client.TransferOutAll(ctx, tcg.TransferOutAllRequest{
		Username:    strings.TrimSpace(request.Username),
		ProductType: request.ProductType,
		ReferenceNo: strings.TrimSpace(request.ReferenceNo),
	})
	response := TCGTransferResponse{
		Status:      result.Status,
		ErrorDesc:   strings.TrimSpace(result.ErrorDesc),
		ReferenceNo: strings.TrimSpace(request.ReferenceNo),
		Raw:         result.RawPayload,
	}
	return response, err
}

func (s *TCGService) CheckTransferStatus(ctx context.Context, request TCGTransferStatusRequest) (TCGTransferStatusResponse, error) {
	if err := s.ensureReady("check_transfer_status", ""); err != nil {
		return TCGTransferStatusResponse{}, err
	}
	result, err := s.client.CheckTransferStatus(ctx, tcg.CheckTransferStatusRequest{
		ProductType: request.ProductType,
		ReferenceNo: strings.TrimSpace(request.ReferenceNo),
	})
	details, _ := result.RawPayload["transaction_details"].(map[string]any)
	response := TCGTransferStatusResponse{
		Status:            result.Status,
		ErrorDesc:         strings.TrimSpace(result.ErrorDesc),
		TransactionStatus: firstNonEmptyTCGValue(result.RawPayload, "transaction_status"),
		TransactionDetail: details,
		Raw:               result.RawPayload,
	}
	return response, err
}

func (s *TCGService) LaunchGame(ctx context.Context, request TCGLaunchRequest) (TCGLaunchResponse, error) {
	if err := s.ensureReady("launch_game", request.Username); err != nil {
		return TCGLaunchResponse{}, err
	}
	result, err := s.client.LaunchGame(ctx, tcg.LaunchGameRequest{
		Username:       strings.TrimSpace(request.Username),
		ProductType:    request.ProductType,
		IPAddress:      strings.TrimSpace(request.IPAddress),
		Platform:       strings.TrimSpace(request.Platform),
		GameMode:       strings.TrimSpace(request.GameMode),
		GameCode:       strings.TrimSpace(request.GameCode),
		Language:       strings.TrimSpace(request.Language),
		Nickname:       strings.TrimSpace(request.Nickname),
		VIPLevel:       request.VIPLevel,
		View:           strings.TrimSpace(request.View),
		BackURL:        strings.TrimSpace(request.BackURL),
		LotteryBetMode: strings.TrimSpace(request.LotteryBetMode),
		Series:         request.Series,
	})
	response := TCGLaunchResponse{
		Status:     result.Status,
		ErrorDesc:  strings.TrimSpace(result.ErrorDesc),
		GameURL:    firstNonEmptyTCGValue(result.RawPayload, "game_url"),
		PTUsername: firstNonEmptyTCGValue(result.RawPayload, "pt_username"),
		PTPassword: firstNonEmptyTCGValue(result.RawPayload, "pt_password"),
		Raw:        result.RawPayload,
	}
	return response, err
}

func (s *TCGService) ensureReady(action string, username string) error {
	if !s.enabled {
		log.Printf("[gate][tcg.service.error] action=%s reason=integration_disabled username=%s", action, strings.TrimSpace(username))
		return fmt.Errorf("tcg integration is disabled")
	}
	if s.client == nil {
		log.Printf("[gate][tcg.service.error] action=%s reason=client_missing username=%s", action, strings.TrimSpace(username))
		return fmt.Errorf("tcg client is not configured")
	}
	return nil
}

func firstNonEmptyTCGValue(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
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

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gate/internal/domain/event"
	ginclient "gate/internal/integration/gin"
	nowpayments "gate/internal/integration/nowpayments"
)

const (
	providerNowPayments          = "nowpayments"
	providerNowPaymentsForGin    = "nowpayments_usdt"
	depositWebhookPathNowPayment = "/v1/webhooks/deposits/nowpayments"
)

type WebhookConfig struct {
	GateInternalToken        string
	PublicBaseURL            string
	NowPaymentsIPNSecret     string
	NowPaymentsPayCurrency   string
	NowPaymentsPriceCurrency string
}

type CreateNowPaymentsDepositRequest struct {
	ClientRef string `json:"client_ref"`
	Amount    string `json:"amount"`
}

type CreateNowPaymentsDepositResponse struct {
	Provider      string         `json:"provider"`
	PaymentID     string         `json:"payment_id"`
	PaymentStatus string         `json:"payment_status"`
	PayAddress    string         `json:"pay_address"`
	PayAmount     string         `json:"pay_amount"`
	PayCurrency   string         `json:"pay_currency"`
	PayinExtraID  string         `json:"payin_extra_id"`
	InvoiceURL    string         `json:"invoice_url"`
	Raw           map[string]any `json:"raw"`
}

type WebhookService struct {
	ginClient     *ginclient.Client
	nowPayments   *nowpayments.Client
	internalToken string
	publicBaseURL string
	ipnSecret     string
	payCurrency   string
	priceCurrency string
}

func NewWebhookService(
	ginClient *ginclient.Client,
	nowPayments *nowpayments.Client,
	config WebhookConfig,
) *WebhookService {
	return &WebhookService{
		ginClient:     ginClient,
		nowPayments:   nowPayments,
		internalToken: strings.TrimSpace(config.GateInternalToken),
		publicBaseURL: strings.TrimRight(strings.TrimSpace(config.PublicBaseURL), "/"),
		ipnSecret:     strings.TrimSpace(config.NowPaymentsIPNSecret),
		payCurrency:   strings.ToLower(strings.TrimSpace(config.NowPaymentsPayCurrency)),
		priceCurrency: strings.ToLower(strings.TrimSpace(config.NowPaymentsPriceCurrency)),
	}
}

func (s *WebhookService) InternalToken() string {
	return s.internalToken
}

func (s *WebhookService) CreateNowPaymentsDeposit(ctx context.Context, request CreateNowPaymentsDepositRequest) (CreateNowPaymentsDepositResponse, error) {
	if s.nowPayments == nil {
		return CreateNowPaymentsDepositResponse{}, fmt.Errorf("nowpayments client is not configured")
	}

	clientRef := strings.TrimSpace(request.ClientRef)
	amount := strings.TrimSpace(request.Amount)
	if clientRef == "" || amount == "" {
		return CreateNowPaymentsDepositResponse{}, fmt.Errorf("client_ref and amount are required")
	}

	payCurrency := s.payCurrency
	if payCurrency == "" {
		payCurrency = "usdttrc20"
	}
	priceCurrency := s.priceCurrency
	if priceCurrency == "" {
		priceCurrency = "usd"
	}

	callbackURL := s.publicBaseURL + depositWebhookPathNowPayment
	if strings.TrimSpace(s.publicBaseURL) == "" {
		callbackURL = ""
	}

	created, err := s.nowPayments.CreatePayment(ctx, nowpayments.CreatePaymentRequest{
		PriceAmount:      amount,
		PriceCurrency:    priceCurrency,
		PayCurrency:      payCurrency,
		OrderID:          clientRef,
		OrderDescription: "deposit " + clientRef,
		IPNCallbackURL:   callbackURL,
	})
	if err != nil {
		return CreateNowPaymentsDepositResponse{}, err
	}

	return CreateNowPaymentsDepositResponse{
		Provider:      providerNowPaymentsForGin,
		PaymentID:     created.PaymentID,
		PaymentStatus: created.PaymentStatus,
		PayAddress:    created.PayAddress,
		PayAmount:     created.PayAmount,
		PayCurrency:   created.PayCurrency,
		PayinExtraID:  created.PayinExtraID,
		InvoiceURL:    created.InvoiceURL,
		Raw:           created.Raw,
	}, nil
}

func (s *WebhookService) HandleDepositWebhook(
	ctx context.Context,
	provider string,
	payload map[string]any,
	rawBody []byte,
	headers http.Header,
) (event.WebhookEvent, error) {
	if provider == "" {
		return event.WebhookEvent{}, fmt.Errorf("provider is required")
	}

	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))
	if normalizedProvider == providerNowPayments {
		if err := s.verifyNowPaymentsSignature(rawBody, headers); err != nil {
			return event.WebhookEvent{}, err
		}
	}

	webhookEvent := event.WebhookEvent{
		Provider:   normalizedProvider,
		Type:       "deposit.callback",
		ReceivedAt: time.Now(),
		Payload:    payload,
	}

	log.Printf("[gate] webhook provider=%s payload=%v", normalizedProvider, payload)

	if s.ginClient != nil {
		request, err := s.buildApplyRequest(normalizedProvider, payload)
		if err != nil {
			return webhookEvent, err
		}
		if err := s.ginClient.ApplyDeposit(ctx, request); err != nil {
			return webhookEvent, err
		}
	}

	return webhookEvent, nil
}

func (s *WebhookService) verifyNowPaymentsSignature(rawBody []byte, headers http.Header) error {
	if s.ipnSecret == "" {
		return fmt.Errorf("nowpayments ipn secret is not configured")
	}

	signature := strings.TrimSpace(headers.Get("x-nowpayments-sig"))
	if signature == "" {
		signature = strings.TrimSpace(headers.Get("X-NowPayments-Sig"))
	}
	if signature == "" {
		return fmt.Errorf("missing nowpayments signature")
	}

	mac := hmac.New(sha512.New, []byte(s.ipnSecret))
	_, _ = mac.Write(rawBody)
	expectedHex := hex.EncodeToString(mac.Sum(nil))

	got, err := hex.DecodeString(strings.ToLower(signature))
	if err != nil {
		return fmt.Errorf("invalid nowpayments signature format")
	}
	expected, _ := hex.DecodeString(expectedHex)

	if !hmac.Equal(got, expected) {
		return fmt.Errorf("invalid nowpayments signature")
	}

	return nil
}

func (s *WebhookService) buildApplyRequest(provider string, payload map[string]any) (event.DepositApplyRequest, error) {
	request := event.DepositApplyRequest{
		Provider:       provider,
		ProviderStatus: firstNonEmptyString(payload, []string{"provider_status", "status", "state", "payment_status", "code"}),
		ClientRef:      firstNonEmptyString(payload, []string{"client_ref", "order_id", "order_code", "orderCode", "reference", "ref"}),
		ProviderTxnID:  firstNonEmptyString(payload, []string{"provider_txn_id", "payment_id", "transaction_id", "transactionId", "txid", "tx_hash", "txHash"}),
		Amount:         firstNonEmptyString(payload, []string{"actually_paid", "pay_amount", "amount", "outcome_amount", "paid_amount", "transfer_amount", "price_amount"}),
		Currency:       strings.ToUpper(firstNonEmptyString(payload, []string{"currency", "pay_currency", "outcome_currency", "price_currency"})),
		PaidAt:         time.Now(),
		Raw:            payload,
	}

	if provider == providerNowPayments {
		request.Provider = providerNowPaymentsForGin
		request.ProviderStatus = normalizeNowPaymentsStatus(request.ProviderStatus)
		if strings.TrimSpace(request.Amount) == "" {
			if request.ProviderStatus == "finished" {
				return event.DepositApplyRequest{}, fmt.Errorf("nowpayments finished event missing amount")
			}
			request.Amount = "0"
		}
		if request.Currency == "" {
			request.Currency = "USDT"
		}
	}

	if request.Currency == "" {
		request.Currency = "VND"
	}

	if request.ClientRef == "" && request.ProviderTxnID == "" {
		return event.DepositApplyRequest{}, fmt.Errorf("client_ref or provider_txn_id is required")
	}

	return request, nil
}

func normalizeNowPaymentsStatus(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "finished":
		return "finished"
	case "failed", "refunded", "expired":
		return normalized
	default:
		return "pending"
	}
}

func firstNonEmptyString(payload map[string]any, keys []string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			trimmed := strings.TrimSpace(fmt.Sprint(value))
			if trimmed != "" && trimmed != "<nil>" {
				return trimmed
			}
		}
	}

	return ""
}

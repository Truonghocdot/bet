package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	providerSepay                = "sepay"
)

// sepayClientRefRe extracts a DEP/WD transaction code from SePay's content string.
// SePay content looks like: "DEPed9e4061cb1320e083ade36e566ad947   Ma giao dich  Trace..."
var sepayClientRefRe = regexp.MustCompile(`(?i)(DEP|WD)[a-zA-Z0-9]+`)

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
	ginClient           *ginclient.Client
	nowPayments         *nowpayments.Client
	credentialsProvider NowPaymentsCredentialsProvider
	internalToken       string
	publicBaseURL       string
	fallbackAPIKey      string
	fallbackIPNSecret   string
	payCurrency         string
	priceCurrency       string
}

func NewWebhookService(
	ginClient *ginclient.Client,
	nowPayments *nowpayments.Client,
	config WebhookConfig,
) *WebhookService {
	return &WebhookService{
		ginClient:         ginClient,
		nowPayments:       nowPayments,
		internalToken:     strings.TrimSpace(config.GateInternalToken),
		publicBaseURL:     strings.TrimRight(strings.TrimSpace(config.PublicBaseURL), "/"),
		fallbackIPNSecret: strings.TrimSpace(config.NowPaymentsIPNSecret),
		payCurrency:       strings.ToLower(strings.TrimSpace(config.NowPaymentsPayCurrency)),
		priceCurrency:     strings.ToLower(strings.TrimSpace(config.NowPaymentsPriceCurrency)),
	}
}

func (s *WebhookService) InternalToken() string {
	return s.internalToken
}

func (s *WebhookService) SetCredentialsProvider(provider NowPaymentsCredentialsProvider) {
	s.credentialsProvider = provider
}

func (s *WebhookService) SetFallbackAPIKey(apiKey string) {
	s.fallbackAPIKey = strings.TrimSpace(apiKey)
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

	credentials, err := s.resolveNowPaymentsCredentials(ctx)
	if err != nil {
		log.Printf("[gate][nowpayments.credentials.warn] source=fallback err=%v", err)
	}

	payCurrency := credentials.PayCurrency
	if payCurrency == "" {
		payCurrency = s.payCurrency
	}
	if payCurrency == "" {
		payCurrency = "usdttrc20"
	}

	priceCurrency := credentials.PriceCurrency
	if priceCurrency == "" {
		priceCurrency = s.priceCurrency
	}
	if priceCurrency == "" {
		priceCurrency = "usd"
	}

	callbackURL := s.publicBaseURL + depositWebhookPathNowPayment
	if strings.TrimSpace(s.publicBaseURL) == "" {
		callbackURL = ""
	}

	created, err := s.nowPayments.CreatePaymentWithAPIKey(ctx, credentials.APIKey, nowpayments.CreatePaymentRequest{
		PriceAmount:      amount,
		PriceCurrency:    priceCurrency,
		PayAmount:        amount,
		PayCurrency:      payCurrency,
		OrderID:          clientRef,
		OrderDescription: "deposit " + clientRef,
		IPNCallbackURL:   callbackURL,
	})
	if err != nil {
		log.Printf(
			"[gate][nowpayments.create.error] client_ref=%s amount=%s price_currency=%s pay_currency=%s source=%s err=%v",
			clientRef,
			amount,
			priceCurrency,
			payCurrency,
			credentials.Source,
			err,
		)
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
		if err := s.verifyNowPaymentsSignature(ctx, rawBody, headers); err != nil {
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
	fmt.Printf("[gate][debug] RECEIVED WEBHOOK provider=%s payload_keys=%d\n", normalizedProvider, len(payload))

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

func (s *WebhookService) verifyNowPaymentsSignature(ctx context.Context, rawBody []byte, headers http.Header) error {
	credentials, err := s.resolveNowPaymentsCredentials(ctx)
	if err != nil {
		log.Printf("[gate][nowpayments.credentials.warn] source=fallback err=%v", err)
	}

	secret := strings.TrimSpace(credentials.IPNSecret)
	if secret == "" {
		return fmt.Errorf("nowpayments ipn secret is not configured")
	}

	signature := strings.TrimSpace(headers.Get("x-nowpayments-sig"))
	if signature == "" {
		signature = strings.TrimSpace(headers.Get("X-NowPayments-Sig"))
	}
	if signature == "" {
		return fmt.Errorf("missing nowpayments signature")
	}

	mac := hmac.New(sha512.New, []byte(secret))
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

func (s *WebhookService) resolveNowPaymentsCredentials(ctx context.Context) (NowPaymentsCredentials, error) {
	fallback := NowPaymentsCredentials{
		APIKey:    strings.TrimSpace(s.fallbackAPIKey),
		IPNSecret: strings.TrimSpace(s.fallbackIPNSecret),
		Source:    "env",
	}

	if s.credentialsProvider == nil {
		return fallback, nil
	}

	credentials, err := s.credentialsProvider.Get(ctx)
	if err != nil {
		return fallback, err
	}

	if strings.TrimSpace(credentials.APIKey) == "" {
		credentials.APIKey = fallback.APIKey
	}
	if strings.TrimSpace(credentials.IPNSecret) == "" {
		credentials.IPNSecret = fallback.IPNSecret
	}
	if strings.TrimSpace(credentials.Source) == "" {
		credentials.Source = fallback.Source
	}

	return credentials, nil
}

func (s *WebhookService) buildApplyRequest(provider string, payload map[string]any) (event.DepositApplyRequest, error) {
	log.Printf("[gate][debug] building apply request for provider=%s", provider)

	// ── SePay-specific handling ──────────────────────────────────────────────
	if provider == providerSepay {
		return s.buildSepayApplyRequest(payload)
	}

	// ── Generic mapping ──────────────────────────────────────────────────────
	request := event.DepositApplyRequest{
		Provider:       provider,
		ProviderStatus: firstNonEmptyString(payload, []string{"provider_status", "status", "state", "payment_status", "code"}),
		ClientRef:      firstNonEmptyString(payload, []string{"client_ref", "order_id", "order_code", "orderCode", "reference", "ref", "content"}),
		ProviderTxnID:  firstNonEmptyString(payload, []string{"provider_txn_id", "payment_id", "transaction_id", "transactionId", "txid", "tx_hash", "txHash", "id"}),
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
		log.Printf("[gate][debug.error] both ClientRef and ProviderTxnID are empty for provider=%s", provider)
		return event.DepositApplyRequest{}, fmt.Errorf("client_ref or provider_txn_id is required (v2.1)")
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

// buildSepayApplyRequest maps SePay's specific webhook payload to a DepositApplyRequest.
//
// SePay payload example:
//
//	{
//	  "gateway": "MBBank",
//	  "content": "DEPed9e4061cb1320e083ade36e566ad947   Ma giao dich  Trace260709",
//	  "transferAmount": 50000,
//	  "transferType": "in",
//	  "referenceCode": "FT26103263302800",
//	  "id": 50026418
//	}
func (s *WebhookService) buildSepayApplyRequest(payload map[string]any) (event.DepositApplyRequest, error) {
	// ── 1. Extract client_ref from the content field via regex ────────────────
	clientRef := ""
	if raw, ok := payload["content"]; ok {
		content := strings.TrimSpace(fmt.Sprint(raw))
		if match := sepayClientRefRe.FindString(content); match != "" {
			clientRef = match
			log.Printf("[gate][sepay] extracted client_ref=%s from content=%q", clientRef, content)
		}
	}

	// ── 2. provider_txn_id: prefer referenceCode, fallback to id ─────────────
	providerTxnID := firstNonEmptyString(payload, []string{"referenceCode", "reference_code", "id"})

	// ── 3. Amount from transferAmount ─────────────────────────────────────────
	amount := firstNonEmptyString(payload, []string{"transferAmount", "transfer_amount", "amount"})

	// ── 4. Status: "in" = finished (money received) ───────────────────────────
	transferType := strings.ToLower(strings.TrimSpace(fmt.Sprint(payload["transferType"])))
	providerStatus := "pending"
	if transferType == "in" {
		providerStatus = "finished"
	}

	log.Printf("[gate][sepay] client_ref=%q provider_txn_id=%q amount=%q status=%s",
		clientRef, providerTxnID, amount, providerStatus)

	if clientRef == "" && providerTxnID == "" {
		return event.DepositApplyRequest{}, fmt.Errorf("sepay: could not extract client_ref from content and no id/referenceCode found")
	}

	return event.DepositApplyRequest{
		Provider:       providerSepay,
		ProviderStatus: providerStatus,
		ClientRef:      clientRef,
		ProviderTxnID:  providerTxnID,
		Amount:         amount,
		Currency:       "VND",
		PaidAt:         time.Now(),
		Raw:            payload,
	}, nil
}

package nowpayments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type CreatePaymentRequest struct {
	PriceAmount      string
	PriceCurrency    string
	PayAmount        string
	PayCurrency      string
	OrderID          string
	OrderDescription string
	IPNCallbackURL   string
}

type CreatePaymentResponse struct {
	PaymentID     string
	PaymentStatus string
	PayAddress    string
	PayAmount     string
	PayCurrency   string
	PayinExtraID  string
	InvoiceURL    string
	Raw           map[string]any
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) CreatePayment(ctx context.Context, request CreatePaymentRequest) (CreatePaymentResponse, error) {
	return c.CreatePaymentWithAPIKey(ctx, c.apiKey, request)
}

func (c *Client) CreatePaymentWithAPIKey(ctx context.Context, apiKey string, request CreatePaymentRequest) (CreatePaymentResponse, error) {
	if c.baseURL == "" {
		return CreatePaymentResponse{}, fmt.Errorf("nowpayments base url is required")
	}
	resolvedAPIKey := strings.TrimSpace(apiKey)
	if resolvedAPIKey == "" {
		return CreatePaymentResponse{}, fmt.Errorf("nowpayments api key is required")
	}

	priceAmount, err := strconv.ParseFloat(strings.TrimSpace(request.PriceAmount), 64)
	if err != nil || priceAmount <= 0 {
		return CreatePaymentResponse{}, fmt.Errorf("invalid price amount")
	}

	bodyMap := map[string]any{
		"price_amount":      priceAmount,
		"price_currency":    strings.ToUpper(strings.TrimSpace(request.PriceCurrency)),
		"pay_currency":      strings.ToUpper(strings.TrimSpace(request.PayCurrency)),
		"order_id":          strings.TrimSpace(request.OrderID),
		"order_description": strings.TrimSpace(request.OrderDescription),
		"ipn_callback_url":  strings.TrimSpace(request.IPNCallbackURL),
	}

	if payAmountStr := strings.TrimSpace(request.PayAmount); payAmountStr != "" {
		if payAmount, err := strconv.ParseFloat(payAmountStr, 64); err == nil && payAmount > 0 {
			bodyMap["pay_amount"] = payAmount
		}
	}

	body, err := json.Marshal(bodyMap)
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/payment", bytes.NewReader(body))
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("x-api-key", resolvedAPIKey)

	httpResponse, err := c.client.Do(httpRequest)
	if err != nil {
		return CreatePaymentResponse{}, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		rawBody, _ := io.ReadAll(httpResponse.Body)
		bodyText := strings.TrimSpace(string(rawBody))
		if bodyText == "" {
			return CreatePaymentResponse{}, fmt.Errorf("nowpayments create payment returned status %d", httpResponse.StatusCode)
		}

		var payload map[string]any
		if err := json.Unmarshal(rawBody, &payload); err == nil {
			if message := firstNonEmptyString(payload, []string{"message", "error", "errors"}); message != "" {
				return CreatePaymentResponse{}, fmt.Errorf(
					"nowpayments create payment returned status %d message=%s body=%s",
					httpResponse.StatusCode,
					message,
					bodyText,
				)
			}
		}

		return CreatePaymentResponse{}, fmt.Errorf(
			"nowpayments create payment returned status %d body=%s",
			httpResponse.StatusCode,
			bodyText,
		)
	}

	var raw map[string]any
	if err := json.NewDecoder(httpResponse.Body).Decode(&raw); err != nil {
		return CreatePaymentResponse{}, err
	}

	response := CreatePaymentResponse{
		PaymentID:     firstNonEmptyString(raw, []string{"payment_id", "id"}),
		PaymentStatus: firstNonEmptyString(raw, []string{"payment_status", "status"}),
		PayAddress:    firstNonEmptyString(raw, []string{"pay_address", "payin_address"}),
		PayAmount:     firstNonEmptyString(raw, []string{"pay_amount", "amount"}),
		PayCurrency:   strings.ToUpper(firstNonEmptyString(raw, []string{"pay_currency", "currency"})),
		PayinExtraID:  firstNonEmptyString(raw, []string{"payin_extra_id", "memo", "destination_tag"}),
		InvoiceURL:    firstNonEmptyString(raw, []string{"invoice_url", "pay_url", "payment_url"}),
		Raw:           raw,
	}

	if response.PaymentID == "" {
		return CreatePaymentResponse{}, fmt.Errorf("nowpayments response missing payment_id")
	}

	return response, nil
}

func firstNonEmptyString(payload map[string]any, keys []string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			var trimmed string
			switch v := value.(type) {
			case float64:
				trimmed = fmt.Sprintf("%.0f", v)
			default:
				trimmed = strings.TrimSpace(fmt.Sprint(value))
			}

			if trimmed != "" && trimmed != "<nil>" {
				return trimmed
			}
		}
	}

	return ""
}

package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DepositClient struct {
	baseURL       string
	internalToken string
	client        *http.Client
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

func NewDepositClient(baseURL, internalToken string) *DepositClient {
	return &DepositClient{
		baseURL:       strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		internalToken: strings.TrimSpace(internalToken),
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
	}
}

func (c *DepositClient) CreateNowPaymentsDeposit(ctx context.Context, request CreateNowPaymentsDepositRequest) (CreateNowPaymentsDepositResponse, error) {
	if c.baseURL == "" {
		return CreateNowPaymentsDepositResponse{}, fmt.Errorf("gate base url is required")
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return CreateNowPaymentsDepositResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/internal/v1/nowpayments/deposits/create",
		bytes.NewReader(payload),
	)
	if err != nil {
		return CreateNowPaymentsDepositResponse{}, err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Internal-Token", c.internalToken)

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return CreateNowPaymentsDepositResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return CreateNowPaymentsDepositResponse{}, fmt.Errorf("gate nowpayments create returned status %d", response.StatusCode)
	}

	var parsed CreateNowPaymentsDepositResponse
	if err := json.NewDecoder(response.Body).Decode(&parsed); err != nil {
		return CreateNowPaymentsDepositResponse{}, err
	}

	return parsed, nil
}

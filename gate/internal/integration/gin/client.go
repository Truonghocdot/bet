package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gate/internal/domain/event"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		token:   strings.TrimSpace(token),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) ApplyDeposit(ctx context.Context, request event.DepositApplyRequest) error {
	if c.baseURL == "" {
		return fmt.Errorf("gin internal base url is required")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/deposits/apply", bytes.NewReader(body))
	if err != nil {
		return err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Internal-Token", c.token)

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("gin internal deposit apply returned status %d", response.StatusCode)
	}

	return nil
}

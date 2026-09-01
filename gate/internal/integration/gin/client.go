package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		bodyBytes, _ := io.ReadAll(response.Body)
		return fmt.Errorf("gin internal deposit apply returned status %d: %s", response.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) LookupDepositForNotification(ctx context.Context, request event.DepositNotificationLookupRequest) (event.DepositNotificationLookup, error) {
	var response event.DepositNotificationLookup
	if err := c.doJSON(ctx, http.MethodPost, "/internal/v1/deposits/lookup-for-notification", request, &response); err != nil {
		return event.DepositNotificationLookup{}, err
	}

	return response, nil
}

func (c *Client) RecordTelegramGroupEvent(ctx context.Context, request event.TelegramGroupEvent) error {
	return c.doJSON(ctx, http.MethodPost, "/internal/v1/telegram/group-events", request, nil)
}

func (c *Client) ListTelegramTargets(ctx context.Context, siteCode string) ([]event.TelegramTarget, error) {
	var response struct {
		Targets []event.TelegramTarget `json:"targets"`
	}
	path := "/internal/v1/telegram/targets?site_code=" + url.QueryEscape(strings.TrimSpace(siteCode))
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}

	return response.Targets, nil
}

func (c *Client) MarkTelegramTargetError(ctx context.Context, request event.TelegramTargetError) error {
	return c.doJSON(ctx, http.MethodPost, "/internal/v1/telegram/target-errors", request, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload any, result any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	return c.doRequest(ctx, method, path, body, result)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, result any) error {
	if c.baseURL == "" {
		return fmt.Errorf("gin internal base url is required")
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	httpRequest.Header.Set("X-Internal-Token", c.token)
	if body != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(response.Body)
		return fmt.Errorf("gin internal request returned status %d: %s", response.StatusCode, string(bodyBytes))
	}
	if result == nil {
		return nil
	}

	return json.NewDecoder(response.Body).Decode(result)
}

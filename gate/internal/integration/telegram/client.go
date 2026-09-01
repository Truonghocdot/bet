package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	token  string
	client *http.Client
}

type APIError struct {
	StatusCode int
	Code       int
	RetryAfter time.Duration
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("telegram api error status=%d code=%d: %s", e.StatusCode, e.Code, e.Message)
}

func NewClient(token string) *Client {
	return &Client{
		token:  strings.TrimSpace(token),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, message string) error {
	if c == nil || c.token == "" {
		return fmt.Errorf("telegram bot token is required")
	}

	body, err := json.Marshal(map[string]any{
		"chat_id":                  chatID,
		"text":                     message,
		"disable_web_page_preview": true,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+c.token+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	var payload struct {
		OK          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code"`
		Description string `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return nil
		}
		return &APIError{StatusCode: response.StatusCode, Message: strings.TrimSpace(string(responseBody))}
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 && payload.OK {
		return nil
	}

	return &APIError{
		StatusCode: response.StatusCode,
		Code:       payload.ErrorCode,
		RetryAfter: time.Duration(payload.Parameters.RetryAfter) * time.Second,
		Message:    strings.TrimSpace(payload.Description),
	}
}

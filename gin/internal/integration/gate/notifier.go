package gate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gin/internal/support/message"
)

type NotificationRequest struct {
	Channel string         `json:"channel"`
	Target  string         `json:"target"`
	Subject string         `json:"subject"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta"`
}

type Notifier struct {
	baseURL string
	client  *http.Client
}

func NewNotifier(baseURL string) *Notifier {
	return &Notifier{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) Send(ctx context.Context, request NotificationRequest) error {
	if n.baseURL == "" {
		return fmt.Errorf(message.GateBaseURLRequired)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		n.baseURL+"/v1/notifications/"+request.Channel,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := n.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("gate notification returned status %d", response.StatusCode)
	}

	return nil
}

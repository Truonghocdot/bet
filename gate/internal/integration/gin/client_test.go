package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gate/internal/domain/event"
)

func TestApplyDepositReturnsBusinessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "test-token" {
			t.Fatalf("unexpected internal token")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"failed"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	result, err := client.ApplyDeposit(context.Background(), event.DepositApplyRequest{
		ClientRef: "DEP-test",
		Provider:  "sepay",
		Amount:    "50000",
	})
	if err != nil {
		t.Fatalf("apply deposit: %v", err)
	}
	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed", result.Status)
	}
}

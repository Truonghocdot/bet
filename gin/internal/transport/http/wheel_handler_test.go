package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	repopg "gin/internal/repository/postgres"
)

func TestWheelConfigurationErrorHasActionableResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := &WheelHandler{}

	handler.writeError(recorder, fmt.Errorf("%w: round_count=4", repopg.ErrWheelRoundConfiguration))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["code"] != "ROUND_CONFIGURATION_INVALID" {
		t.Fatalf("code = %q", payload["code"])
	}
}

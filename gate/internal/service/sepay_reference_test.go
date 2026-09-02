package service

import (
	"testing"
)

func TestBuildSepayApplyRequestAcceptsReadableManualCode(t *testing.T) {
	request, err := (&WebhookService{}).buildSepayApplyRequest(map[string]any{
		"content":        "Chuyen khoan FHABCDEFGH23",
		"referenceCode":  "FT-1",
		"transferAmount": float64(200000),
		"transferType":   "in",
	})
	if err != nil {
		t.Fatalf("build SePay request: %v", err)
	}
	if request.ClientRef != "DEP-FHABCDEFGH23" {
		t.Fatalf("client_ref = %q, want DEP-FHABCDEFGH23", request.ClientRef)
	}
}

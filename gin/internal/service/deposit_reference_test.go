package service

import (
	"strings"
	"testing"

	repopg "gin/internal/repository/postgres"
)

func TestNewReadableDepositClientRefIsShortAndManualFriendly(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		clientRef := newReadableDepositClientRef("fh88u")
		visible := readableDepositClientRef(clientRef)
		if len(visible) != 12 || !strings.HasPrefix(visible, "FH") {
			t.Fatalf("unexpected readable deposit code %q", visible)
		}
		for _, character := range visible {
			if strings.ContainsRune("01IO", character) {
				t.Fatalf("readable deposit code contains ambiguous character %q: %s", character, visible)
			}
		}
		if _, exists := seen[visible]; exists {
			t.Fatalf("duplicate readable deposit code generated: %s", visible)
		}
		seen[visible] = struct{}{}
	}
}

func TestVietQrPayloadUsesVisibleDepositCode(t *testing.T) {
	clientRef := "DEP-FHABCDEFGH23"
	payload := buildQRContent(repopg.ReceivingAccountRecord{
		ProviderCode:  stringPtrForTest("MBBank"),
		AccountNumber: stringPtrForTest("0327182537"),
	}, "200000", readableDepositClientRef(clientRef))

	if !strings.HasSuffix(payload, "|FHABCDEFGH23") {
		t.Fatalf("QR payload should contain the short visible code, got %q", payload)
	}
}

func stringPtrForTest(value string) *string {
	return &value
}

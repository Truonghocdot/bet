package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"gate/internal/domain/event"
)

func TestSepayNotifyOnlyDoesNotApplyDeposit(t *testing.T) {
	service := NewWebhookService(nil, nil, WebhookConfig{SepayAutoApply: false})

	_, err := service.HandleDepositWebhook(context.Background(), "sepay", map[string]any{
		"id":             float64(50026418),
		"content":        "DEPabc12345",
		"transferAmount": float64(50000),
		"transferType":   "in",
	}, nil, http.Header{})
	if err != nil {
		t.Fatalf("notify-only SePay webhook returned error: %v", err)
	}
}

func TestFormatDepositMessageIncludesManualApprovalAndUnmatchedWarning(t *testing.T) {
	notification := depositNotificationEvent{
		ProviderTxnID: "FT26103263302800",
		ClientRef:     "DEP-abc12345",
		Amount:        "50000",
		Content:       "DEPabc12345",
		PaidAt:        time.Date(2026, time.September, 1, 7, 30, 12, 0, time.UTC),
		Lookup: event.DepositNotificationLookup{
			Matched:          true,
			UserID:           123,
			UserName:         "Nguyen Van A",
			UserPhone:        "0900000000",
			ClientRef:        "DEP-abc12345",
			Amount:           "50000",
			Status:           1,
			ReceivingBank:    "MBBank",
			ReceivingAccount: "0123456789",
		},
	}

	matched := formatDepositMessage(notification)
	for _, expected := range []string{"[NẠP TIỀN SEPAY]", "50.000 VND", "#123", "CHỜ DUYỆT THỦ CÔNG"} {
		if !contains(matched, expected) {
			t.Fatalf("matched message missing %q: %s", expected, matched)
		}
	}

	notification.Lookup = event.DepositNotificationLookup{}
	unmatched := formatDepositMessage(notification)
	for _, expected := range []string{"[CẢNH BÁO TIỀN VÀO CHƯA KHỚP]", "Mã SePay: FT26103263302800", "Không tìm thấy lệnh nạp"} {
		if !contains(unmatched, expected) {
			t.Fatalf("unmatched message missing %q: %s", expected, unmatched)
		}
	}
}

func contains(value, expected string) bool {
	for index := 0; index+len(expected) <= len(value); index++ {
		if value[index:index+len(expected)] == expected {
			return true
		}
	}
	return false
}

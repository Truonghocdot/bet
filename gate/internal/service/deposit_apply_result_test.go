package service

import "testing"

func TestDepositNotificationStatusFromApplyResult(t *testing.T) {
	testCases := map[string]struct {
		status string
		want   depositNotificationCompletionStatus
	}{
		"completed": {status: "completed", want: depositNotificationAutoCompleted},
		"failed":    {status: "failed", want: depositNotificationAutoFailed},
		"pending":   {status: "pending", want: depositNotificationAutoPending},
		"unknown":   {status: "unknown", want: depositNotificationAutoPending},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := depositNotificationStatusFromApplyResult(testCase.status); got != testCase.want {
				t.Fatalf("status %q mapped to %q, want %q", testCase.status, got, testCase.want)
			}
		})
	}
}

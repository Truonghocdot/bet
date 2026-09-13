package service

import (
	"context"
	"testing"
)

type staticNowPaymentsCredentialsProvider struct {
	credentials NowPaymentsCredentials
	err         error
}

func (p staticNowPaymentsCredentialsProvider) Get(context.Context) (NowPaymentsCredentials, error) {
	return p.credentials, p.err
}

func TestSepayAutoApplyUsesRuntimeSnapshot(t *testing.T) {
	runtimeValue := true
	service := NewWebhookService(nil, nil, WebhookConfig{SepayAutoApply: false})
	service.SetCredentialsProvider(staticNowPaymentsCredentialsProvider{
		credentials: NowPaymentsCredentials{SepayAutoApply: &runtimeValue},
	})

	if !service.isSepayAutoApplyEnabled(context.Background(), "1") {
		t.Fatal("expected runtime snapshot to enable automatic SePay deposits")
	}
}

func TestSepayAutoApplyRuntimeSnapshotCanDisableEnvironmentValue(t *testing.T) {
	runtimeValue := false
	service := NewWebhookService(nil, nil, WebhookConfig{SepayAutoApply: true})
	service.SetCredentialsProvider(staticNowPaymentsCredentialsProvider{
		credentials: NowPaymentsCredentials{SepayAutoApply: &runtimeValue},
	})

	if service.isSepayAutoApplyEnabled(context.Background(), "1") {
		t.Fatal("expected runtime snapshot to disable automatic SePay deposits")
	}
}

func TestSepayAutoApplyFallsBackToEnvironmentValue(t *testing.T) {
	service := NewWebhookService(nil, nil, WebhookConfig{SepayAutoApply: true})
	service.SetCredentialsProvider(staticNowPaymentsCredentialsProvider{})

	if !service.isSepayAutoApplyEnabled(context.Background(), "1") {
		t.Fatal("expected environment value when runtime snapshot has no setting")
	}
}

func TestSepayAutoApplyRequiresConfiguredMinimumAmount(t *testing.T) {
	service := NewWebhookService(nil, nil, WebhookConfig{SepayAutoApply: false})
	service.SetCredentialsProvider(staticNowPaymentsCredentialsProvider{
		credentials: NowPaymentsCredentials{
			SepayAutoApply:          boolPointer(true),
			SepayAutoApplyMinAmount: "50000",
		},
	})

	if service.isSepayAutoApplyEnabled(context.Background(), "49999.99") {
		t.Fatal("expected payment below configured minimum to remain pending")
	}
	if !service.isSepayAutoApplyEnabled(context.Background(), "50000") {
		t.Fatal("expected payment at configured minimum to complete automatically")
	}
	if service.isSepayAutoApplyEnabled(context.Background(), "invalid") {
		t.Fatal("expected invalid payment amount to remain pending")
	}
}

func boolPointer(value bool) *bool {
	return &value
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type NowPaymentsCredentials struct {
	APIKey        string
	IPNSecret     string
	PayCurrency   string
	PriceCurrency string
	Source        string
}

type NowPaymentsCredentialsProvider interface {
	Get(ctx context.Context) (NowPaymentsCredentials, error)
}

type RedisNowPaymentsCredentialsProvider struct {
	redis    *goredis.Client
	redisKey string
	fallback NowPaymentsCredentials
}

func NewRedisNowPaymentsCredentialsProvider(
	redis *goredis.Client,
	redisKey string,
	fallback NowPaymentsCredentials,
) *RedisNowPaymentsCredentialsProvider {
	return &RedisNowPaymentsCredentialsProvider{
		redis:    redis,
		redisKey: strings.TrimSpace(redisKey),
		fallback: fallback,
	}
}

func (p *RedisNowPaymentsCredentialsProvider) Get(ctx context.Context) (NowPaymentsCredentials, error) {
	creds := p.fallback
	if strings.TrimSpace(creds.Source) == "" {
		creds.Source = "env"
	}

	if p.redis == nil || p.redisKey == "" {
		return creds, nil
	}

	raw, err := p.redis.Get(ctx, p.redisKey).Result()
	if err != nil {
		if err == goredis.Nil {
			return creds, nil
		}
		return creds, fmt.Errorf("read redis key %s failed: %w", p.redisKey, err)
	}

	var payload struct {
		NowPaymentsAPIKey        string `json:"nowpayments_api_key"`
		NowPaymentsIPNSecret     string `json:"nowpayments_ipn_secret"`
		NowPaymentsPayCurrency   string `json:"nowpayments_pay_currency"`
		NowPaymentsPriceCurrency string `json:"nowpayments_price_currency"`
		BaseCurrency             string `json:"base_currency"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return creds, fmt.Errorf("decode redis key %s failed: %w", p.redisKey, err)
	}

	apiKey := strings.TrimSpace(payload.NowPaymentsAPIKey)
	ipnSecret := strings.TrimSpace(payload.NowPaymentsIPNSecret)

	if apiKey != "" {
		creds.APIKey = apiKey
		creds.Source = "redis"
	}
	if ipnSecret != "" {
		creds.IPNSecret = ipnSecret
		if creds.Source != "redis" {
			creds.Source = "redis_partial"
		}
	}

	payCurrency := strings.TrimSpace(payload.NowPaymentsPayCurrency)
	if payCurrency == "" {
		payCurrency = strings.TrimSpace(payload.BaseCurrency)
	}
	if payCurrency != "" {
		creds.PayCurrency = strings.ToUpper(payCurrency)
	}

	priceCurrency := strings.TrimSpace(payload.NowPaymentsPriceCurrency)
	if priceCurrency == "" {
		// If base is USDT, price should probably be USDT too to avoid conversion drift
		if strings.ToUpper(payload.BaseCurrency) == "USDT" {
			priceCurrency = "USD"
		}
	}
	if priceCurrency != "" {
		creds.PriceCurrency = strings.ToUpper(priceCurrency)
	}

	log.Printf("[gate][nowpayments.credentials] source=%s api_key_present=%v ipn_secret_present=%v", creds.Source, creds.APIKey != "", creds.IPNSecret != "")

	return creds, nil
}

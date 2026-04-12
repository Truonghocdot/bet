package app

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName        string
	HTTPAddr           string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ShutdownTimout     time.Duration
	GinInternalBaseURL string
	GinInternalToken   string
	GateInternalToken  string
	PublicBaseURL      string
	NowPaymentsBaseURL string
	NowPaymentsAPIKey  string
	NowPaymentsIPNKey  string
	NowPaymentsPayCode string
	NowPaymentsPrice   string
}

func LoadConfig() Config {
	return Config{
		ServiceName:        getEnv("APP_NAME", "gate"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8082"),
		ReadTimeout:        getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:       getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimout:     getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		GinInternalBaseURL: getEnv("GIN_INTERNAL_BASE_URL", "http://localhost:8081"),
		GinInternalToken:   getEnv("GIN_INTERNAL_TOKEN", ""),
		GateInternalToken:  getEnv("GATE_INTERNAL_TOKEN", getEnv("GIN_INTERNAL_TOKEN", "")),
		PublicBaseURL:      getEnv("GATE_PUBLIC_BASE_URL", "http://localhost:8082"),
		NowPaymentsBaseURL: getEnv("NOWPAYMENTS_BASE_URL", "https://api.nowpayments.io"),
		NowPaymentsAPIKey:  getEnv("NOWPAYMENTS_API_KEY", ""),
		NowPaymentsIPNKey:  getEnv("NOWPAYMENTS_IPN_SECRET", ""),
		NowPaymentsPayCode: getEnv("NOWPAYMENTS_PAY_CURRENCY", "usdttrc20"),
		NowPaymentsPrice:   getEnv("NOWPAYMENTS_PRICE_CURRENCY", "usd"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}

package app

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName          string
	HTTPAddr             string
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	ShutdownTimout       time.Duration
	GinInternalBaseURL   string
	GinInternalToken     string
	GateInternalToken    string
	SharedRedisAddr      string
	SharedRedisPass      string
	SharedRedisDB        int
	ExchangeRateRedisKey string
	PublicBaseURL        string
	NowPaymentsBaseURL   string
	NowPaymentsAPIKey    string
	NowPaymentsIPNKey    string
	NowPaymentsPayCode   string
	NowPaymentsPrice     string
}

func LoadConfig() Config {
	loadEnvFiles(".env", "../.env", "../../.env")

	return Config{
		ServiceName:          getEnv("APP_NAME", "gate"),
		HTTPAddr:             getEnv("HTTP_ADDR", ":8082"),
		ReadTimeout:          getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:         getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimout:       getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		GinInternalBaseURL:   getEnv("GIN_INTERNAL_BASE_URL", "http://localhost:8081"),
		GinInternalToken:     getEnv("GIN_INTERNAL_TOKEN", ""),
		GateInternalToken:    getEnv("GATE_INTERNAL_TOKEN", getEnv("GIN_INTERNAL_TOKEN", "")),
		SharedRedisAddr:      getEnv("SHARED_REDIS_ADDR", getEnv("REDIS_ADDR", "127.0.0.1:6379")),
		SharedRedisPass:      getEnv("SHARED_REDIS_PASSWORD", getEnv("REDIS_PASSWORD", "")),
		SharedRedisDB:        getEnvInt("SHARED_REDIS_DB", getEnvInt("REDIS_DB", 2)),
		ExchangeRateRedisKey: getEnv("EXCHANGE_RATE_REDIS_KEY", "shared:exchange-rate:usdt-vnd"),
		PublicBaseURL:        getEnv("GATE_PUBLIC_BASE_URL", "http://localhost:8082"),
		NowPaymentsBaseURL:   getEnv("NOWPAYMENTS_BASE_URL", "https://api.nowpayments.io"),
		NowPaymentsAPIKey:    getEnv("NOWPAYMENTS_API_KEY", ""),
		NowPaymentsIPNKey:    getEnv("NOWPAYMENTS_IPN_SECRET", ""),
		NowPaymentsPayCode:   getEnv("NOWPAYMENTS_PAY_CURRENCY", "usdttrc20"),
		NowPaymentsPrice:     getEnv("NOWPAYMENTS_PRICE_CURRENCY", "usd"),
	}
}

func loadEnvFiles(paths ...string) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.HasPrefix(line, "export ") {
				line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
			}

			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}

			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			value = strings.Trim(value, `"'`)

			if key == "" {
				continue
			}

			if os.Getenv(key) == "" {
				_ = os.Setenv(key, value)
			}
		}

		_ = file.Close()
		break
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

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

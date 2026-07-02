package app

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName             string
	HTTPAddr                string
	ReadTimeout             time.Duration
	WriteTimeout            time.Duration
	ShutdownTimout          time.Duration
	GinInternalBaseURL      string
	GinInternalToken        string
	GateInternalToken       string
	SharedRedisAddr         string
	SharedRedisPass         string
	SharedRedisDB           int
	ExchangeRateRedisKey    string
	PublicBaseURL           string
	NowPaymentsBaseURL      string
	NowPaymentsAPIKey       string
	NowPaymentsIPNKey       string
	NowPaymentsPayCode      string
	NowPaymentsPrice        string
	TCGEnabled              bool
	TCGBaseURL              string
	TCGHTTPTimeout          time.Duration
	TCGMerchantCode         string
	TCGMerchantDESKey       string
	TCGMerchantSignKey      string
	TCGReportFTPHost        string
	TCGReportFTPPort        int
	TCGReportFTPUsername    string
	TCGReportFTPPassword    string
	TCGReportFTPBaseDir     string
	TCGGameListSyncEnabled  bool
	TCGGameListSyncInterval time.Duration
	TCGGameListRedisKey     string
	TCGGameListProductTypes []int
	TCGGameListPlatform     string
	TCGGameListClientType   string
	TCGGameListTypes        []string
	TCGGameListTypeProducts map[string][]int
	TCGGameListLanguage     string
	TCGGameListPage         int
	TCGGameListPageSize     int
}

func LoadConfig() Config {
	loadEnvFiles(".env", "../.env", "../../.env")

	return Config{
		ServiceName:             getEnv("APP_NAME", "gate"),
		HTTPAddr:                getEnv("HTTP_ADDR", ":8082"),
		ReadTimeout:             getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:            getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimout:          getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		GinInternalBaseURL:      getEnv("GIN_INTERNAL_BASE_URL", "http://localhost:8081"),
		GinInternalToken:        getEnv("GIN_INTERNAL_TOKEN", ""),
		GateInternalToken:       getEnv("GATE_INTERNAL_TOKEN", getEnv("GIN_INTERNAL_TOKEN", "")),
		SharedRedisAddr:         getEnv("SHARED_REDIS_ADDR", getEnv("REDIS_ADDR", "127.0.0.1:6379")),
		SharedRedisPass:         getEnv("SHARED_REDIS_PASSWORD", getEnv("REDIS_PASSWORD", "")),
		SharedRedisDB:           getEnvInt("SHARED_REDIS_DB", getEnvInt("REDIS_DB", 2)),
		ExchangeRateRedisKey:    getEnv("EXCHANGE_RATE_REDIS_KEY", "shared:exchange-rate:usdt-vnd"),
		PublicBaseURL:           getEnv("GATE_PUBLIC_BASE_URL", "http://localhost:8082"),
		NowPaymentsBaseURL:      getEnv("NOWPAYMENTS_BASE_URL", "https://api.nowpayments.io"),
		NowPaymentsAPIKey:       getEnv("NOWPAYMENTS_API_KEY", ""),
		NowPaymentsIPNKey:       getEnv("NOWPAYMENTS_IPN_SECRET", ""),
		NowPaymentsPayCode:      getEnv("NOWPAYMENTS_PAY_CURRENCY", "usdttrc20"),
		NowPaymentsPrice:        getEnv("NOWPAYMENTS_PRICE_CURRENCY", "usd"),
		TCGEnabled:              getEnvBool("TCG_ENABLED", false),
		TCGBaseURL:              getEnv("TCG_BASE_URL", ""),
		TCGHTTPTimeout:          getEnvDuration("TCG_HTTP_TIMEOUT", 30*time.Second),
		TCGMerchantCode:         getEnv("TCG_MERCHANT_CODE", ""),
		TCGMerchantDESKey:       getEnv("TCG_MERCHANT_DES_KEY", ""),
		TCGMerchantSignKey:      getEnv("TCG_MERCHANT_SIGN_KEY", ""),
		TCGReportFTPHost:        getEnv("TCG_REPORT_FTP_HOST", ""),
		TCGReportFTPPort:        getEnvInt("TCG_REPORT_FTP_PORT", 21),
		TCGReportFTPUsername:    getEnv("TCG_REPORT_FTP_USERNAME", ""),
		TCGReportFTPPassword:    getEnv("TCG_REPORT_FTP_PASSWORD", ""),
		TCGReportFTPBaseDir:     getEnv("TCG_REPORT_FTP_BASE_DIR", ""),
		TCGGameListSyncEnabled:  getEnvBool("TCG_GAME_LIST_SYNC_ENABLED", false),
		TCGGameListSyncInterval: getEnvDuration("TCG_GAME_LIST_SYNC_INTERVAL", 5*time.Minute),
		TCGGameListRedisKey:     getEnv("TCG_GAME_LIST_REDIS_KEY", "shared:tcg:game-list:v1"),
		TCGGameListProductTypes: getEnvIntSlice("TCG_GAME_LIST_PRODUCT_TYPES", getEnv("TCG_GAME_LIST_PRODUCT_TYPE", "7")),
		TCGGameListPlatform:     getEnv("TCG_GAME_LIST_PLATFORM", "all"),
		TCGGameListClientType:   getEnv("TCG_GAME_LIST_CLIENT_TYPE", "all"),
		TCGGameListTypes:        getEnvStringSlice("TCG_GAME_LIST_TYPES", getEnv("TCG_GAME_LIST_TYPE", "RNG")),
		TCGGameListTypeProducts: getEnvIntMapByGameType(),
		TCGGameListLanguage:     getEnv("TCG_GAME_LIST_LANGUAGE", ""),
		TCGGameListPage:         getEnvInt("TCG_GAME_LIST_PAGE", 0),
		TCGGameListPageSize:     getEnvInt("TCG_GAME_LIST_PAGE_SIZE", 0),
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

			// Strip inline comments: VALUE   # comment → VALUE
			// Only strip if '#' is preceded by whitespace to avoid breaking tokens like "pass#123"
			if idx := strings.Index(value, " #"); idx != -1 {
				value = strings.TrimSpace(value[:idx])
				value = strings.Trim(value, `"'`)
			}

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

	if duration, err := time.ParseDuration(value); err == nil && duration > 0 {
		return duration
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

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func getEnvIntSlice(key string, fallbackRaw string) []int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		raw = strings.TrimSpace(fallbackRaw)
	}
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	items := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			continue
		}

		items = append(items, parsed)
	}

	return items
}

func getEnvStringSlice(key string, fallbackRaw string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		raw = strings.TrimSpace(fallbackRaw)
	}
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		items = append(items, trimmed)
	}

	return items
}

func getEnvIntMapByGameType() map[string][]int {
	result := map[string][]int{}
	for _, gameType := range []string{"RNG", "FISH", "LIVE", "PVP", "SPORTS", "ELOTT"} {
		key := "TCG_GAME_LIST_PRODUCTS_" + gameType
		items := getEnvIntSlice(key, "")
		if len(items) == 0 {
			continue
		}
		result[gameType] = items
	}
	return result
}

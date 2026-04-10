package app

import (
	"bufio"
	"os"
	"strings"
	"strconv"
	"time"
)

type Config struct {
	ServiceName                      string
	HTTPAddr                         string
	ReadTimeout                      time.Duration
	WriteTimeout                     time.Duration
	ShutdownTimout                   time.Duration
	DatabaseURL                      string
	AuthSecret                       string
	AuthTTL                          time.Duration
	RegisterURL                      string
	RedisAddr                        string
	RedisPassword                    string
	RedisDB                          int
	GateBaseURL                      string
	InternalToken                    string
	PaymentReceivingAccountsRedisKey string
	ForgotOTPTTL                     time.Duration
	ForgotCooldown                   time.Duration
	ForgotMaxTry                     int
	ForgotWindow                     time.Duration
	ForgotLimitIP                    int
	ForgotLimitTarget                int
	LoginFailWindow                  time.Duration
	LoginFailLimitIP                 int
	LoginFailLimitAccount            int
	LoginLockDuration                time.Duration
	RegisterWindow                   time.Duration
	RegisterLimitIP                  int
	RegisterLimitEmail               int
	RegisterLimitPhone               int
}

func LoadConfig() Config {
	loadEnvFiles(".env", "../.env", "../../.env")

	return Config{
		ServiceName:                      getEnv("APP_NAME", "gin-core"),
		HTTPAddr:                         getEnv("HTTP_ADDR", ":8081"),
		ReadTimeout:                      getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:                     getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimout:                   getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		DatabaseURL:                      getEnv("DATABASE_URL", ""),
		AuthSecret:                       getEnv("AUTH_TOKEN_SECRET", ""),
		AuthTTL:                          getEnvDuration("AUTH_TOKEN_TTL", 24*time.Hour),
		RegisterURL:                      getEnv("PUBLIC_REGISTER_URL", "http://localhost:3000/register"),
		RedisAddr:                        getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:                    getEnv("REDIS_PASSWORD", ""),
		RedisDB:                          getEnvInt("REDIS_DB", 0),
		GateBaseURL:                      getEnv("GATE_BASE_URL", "http://localhost:8082"),
		InternalToken:                    getEnv("GIN_INTERNAL_TOKEN", ""),
		PaymentReceivingAccountsRedisKey: getEnv("PAYMENT_RECEIVING_ACCOUNTS_REDIS_KEY", "shared:payment:receiving-accounts:v1"),
		ForgotOTPTTL:                     getEnvDuration("AUTH_FORGOT_OTP_TTL", 5*time.Minute),
		ForgotCooldown:                   getEnvDuration("AUTH_FORGOT_OTP_COOLDOWN", 60*time.Second),
		ForgotMaxTry:                     getEnvInt("AUTH_FORGOT_OTP_MAX_ATTEMPTS", 5),
		ForgotWindow:                     getEnvDuration("AUTH_FORGOT_WINDOW", time.Hour),
		ForgotLimitIP:                    getEnvInt("AUTH_FORGOT_LIMIT_IP", 5),
		ForgotLimitTarget:                getEnvInt("AUTH_FORGOT_LIMIT_TARGET", 3),
		LoginFailWindow:                  getEnvDuration("AUTH_LOGIN_FAIL_WINDOW", 15*time.Minute),
		LoginFailLimitIP:                 getEnvInt("AUTH_LOGIN_FAIL_LIMIT_IP", 10),
		LoginFailLimitAccount:            getEnvInt("AUTH_LOGIN_FAIL_LIMIT_ACCOUNT", 5),
		LoginLockDuration:                getEnvDuration("AUTH_LOGIN_LOCK_DURATION", 15*time.Minute),
		RegisterWindow:                   getEnvDuration("AUTH_REGISTER_WINDOW", 15*time.Minute),
		RegisterLimitIP:                  getEnvInt("AUTH_REGISTER_LIMIT_IP", 5),
		RegisterLimitEmail:               getEnvInt("AUTH_REGISTER_LIMIT_EMAIL", 3),
		RegisterLimitPhone:               getEnvInt("AUTH_REGISTER_LIMIT_PHONE", 3),
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

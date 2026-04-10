package app

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName    string
	HTTPAddr       string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ShutdownTimout time.Duration
}

func LoadConfig() Config {
	return Config{
		ServiceName:    getEnv("APP_NAME", "gate"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8082"),
		ReadTimeout:    getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:   getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimout: getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
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

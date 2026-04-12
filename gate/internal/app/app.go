package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	ginclient "gate/internal/integration/gin"
	nowpayments "gate/internal/integration/nowpayments"
	"gate/internal/service"
	httptransport "gate/internal/transport/http"
	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	config Config
	server *http.Server
	redis  *goredis.Client
}

func New() (*App, error) {
	config := LoadConfig()
	ginClient := ginclient.NewClient(config.GinInternalBaseURL, config.GinInternalToken)
	nowPaymentsClient := nowpayments.NewClient(config.NowPaymentsBaseURL, config.NowPaymentsAPIKey)
	sharedRedis := goredis.NewClient(&goredis.Options{
		Addr:     config.SharedRedisAddr,
		Password: config.SharedRedisPass,
		DB:       config.SharedRedisDB,
	})
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
	pingErr := sharedRedis.Ping(pingCtx).Err()
	cancelPing()
	if pingErr != nil {
		log.Printf("[gate][redis.warn] addr=%s db=%d err=%v", config.SharedRedisAddr, config.SharedRedisDB, pingErr)
		_ = sharedRedis.Close()
		sharedRedis = nil
	}
	webhookService := service.NewWebhookService(ginClient, nowPaymentsClient, service.WebhookConfig{
		GateInternalToken:        config.GateInternalToken,
		PublicBaseURL:            config.PublicBaseURL,
		NowPaymentsIPNSecret:     config.NowPaymentsIPNKey,
		NowPaymentsPayCurrency:   config.NowPaymentsPayCode,
		NowPaymentsPriceCurrency: config.NowPaymentsPrice,
	})
	webhookService.SetFallbackAPIKey(config.NowPaymentsAPIKey)
	webhookService.SetCredentialsProvider(
		service.NewRedisNowPaymentsCredentialsProvider(
			sharedRedis,
			config.ExchangeRateRedisKey,
			service.NowPaymentsCredentials{
				APIKey:    config.NowPaymentsAPIKey,
				IPNSecret: config.NowPaymentsIPNKey,
				Source:    "env",
			},
		),
	)
	notificationService := service.NewNotificationService()
	router := httptransport.NewRouter(webhookService, notificationService)

	server := &http.Server{
		Addr:         config.HTTPAddr,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &App{
		config: config,
		server: server,
		redis:  sharedRedis,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("[%s] listening on %s", a.config.ServiceName, a.config.HTTPAddr)
		serverErr <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimout)
		defer cancel()
		err := a.server.Shutdown(shutdownCtx)
		if a.redis != nil {
			_ = a.redis.Close()
		}
		return err
	case err := <-serverErr:
		if a.redis != nil {
			_ = a.redis.Close()
		}
		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}
}

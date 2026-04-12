package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	ginclient "gate/internal/integration/gin"
	nowpayments "gate/internal/integration/nowpayments"
	"gate/internal/service"
	httptransport "gate/internal/transport/http"
)

type App struct {
	config Config
	server *http.Server
}

func New() (*App, error) {
	config := LoadConfig()
	ginClient := ginclient.NewClient(config.GinInternalBaseURL, config.GinInternalToken)
	nowPaymentsClient := nowpayments.NewClient(config.NowPaymentsBaseURL, config.NowPaymentsAPIKey)
	webhookService := service.NewWebhookService(ginClient, nowPaymentsClient, service.WebhookConfig{
		GateInternalToken:        config.GateInternalToken,
		PublicBaseURL:            config.PublicBaseURL,
		NowPaymentsIPNSecret:     config.NowPaymentsIPNKey,
		NowPaymentsPayCurrency:   config.NowPaymentsPayCode,
		NowPaymentsPriceCurrency: config.NowPaymentsPrice,
	})
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

		return a.server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}
}

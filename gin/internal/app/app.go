package app

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"gin/internal/auth/token"
	"gin/internal/event/outbox"
	"gin/internal/integration/gate"
	platformpg "gin/internal/platform/postgres"
	platformredis "gin/internal/platform/redis"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/security/ratelimit"
	"gin/internal/service"
	httptransport "gin/internal/transport/http"
	"gin/internal/ws"

	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	config Config
	server *http.Server
	db     *sql.DB
	redis  *goredis.Client
}

func New() (*App, error) {
	config := LoadConfig()
	db, err := platformpg.Open(config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	platformpg.ConfigurePool(db, platformpg.PoolConfig{
		MaxOpenConns:    config.DBMaxOpenConns,
		MaxIdleConns:    config.DBMaxIdleConns,
		ConnMaxLifetime: config.DBConnMaxLifetime,
		ConnMaxIdleTime: config.DBConnMaxIdleTime,
	})
	log.Printf("[db.pool] max_open=%d max_idle=%d conn_max_lifetime=%s conn_max_idle_time=%s", config.DBMaxOpenConns, config.DBMaxIdleConns, config.DBConnMaxLifetime, config.DBConnMaxIdleTime)

	tokenSigner, err := token.NewSigner(config.AuthSecret, config.AuthTTL)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	redisClient, err := platformredis.Open(context.Background(), config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	broker := realtime.NewBroker(redisClient)

	hub := ws.NewHub()
	publisher := outbox.NewNoopPublisher()
	userRepository := repopg.NewUserRepository(db)
	walletRepository := repopg.NewWalletRepository(db)
	notificationRepository := repopg.NewNotificationRepository(db)
	gameRepository := repopg.NewGameRepository(db)
	depositRepository := repopg.NewDepositRepository(db)
	withdrawalRepository := repopg.NewWithdrawalRepository(db)
	limiter := ratelimit.New(redisClient)
	notifier := gate.NewNotifier(config.GateBaseURL)
	depositGateway := gate.NewDepositClient(config.GateBaseURL, config.GateInternalToken)
	authService := service.NewAuthService(userRepository, tokenSigner, limiter, notifier, service.AuthConfig{
		RegisterURL:           config.RegisterURL,
		OTPSecret:             config.AuthSecret,
		ForgotOTPTTL:          config.ForgotOTPTTL,
		ForgotCooldown:        config.ForgotCooldown,
		ForgotMaxAttempts:     config.ForgotMaxTry,
		ForgotWindow:          config.ForgotWindow,
		ForgotLimitIP:         config.ForgotLimitIP,
		ForgotLimitTarget:     config.ForgotLimitTarget,
		LoginFailWindow:       config.LoginFailWindow,
		LoginFailLimitIP:      config.LoginFailLimitIP,
		LoginFailLimitAccount: config.LoginFailLimitAccount,
		LoginLockDuration:     config.LoginLockDuration,
		RegisterWindow:        config.RegisterWindow,
		RegisterLimitIP:       config.RegisterLimitIP,
		RegisterLimitEmail:    config.RegisterLimitEmail,
		RegisterLimitPhone:    config.RegisterLimitPhone,
	})
	walletService := service.NewWalletService(walletRepository, broker)
	notificationService := service.NewNotificationService(notificationRepository)
	sessionService := service.NewGameSessionService(hub, walletRepository)
	betService := service.NewBetService(publisher, sessionService, gameRepository, walletRepository)
	playRoomService := service.NewPlayRoomService(gameRepository, walletRepository, walletService, redisClient, broker)
	depositService := service.NewDepositService(depositRepository, redisClient, walletService, depositGateway, service.DepositConfig{
		ReceivingAccountsRedisKey: config.PaymentReceivingAccountsRedisKey,
	})
	withdrawalService := service.NewWithdrawalService(withdrawalRepository, walletRepository)
	router := httptransport.NewRouter(config, authService, walletService, notificationService, sessionService, betService, playRoomService, depositService, withdrawalService, broker, config.InternalToken)

	server := &http.Server{
		Addr:        config.HTTPAddr,
		Handler:     router,
		ReadTimeout: config.ReadTimeout,
		// SSE streams need a long-lived write window.
		WriteTimeout: 0,
	}

	return &App{
		config: config,
		server: server,
		db:     db,
		redis:  redisClient,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("[%s] đang lắng nghe tại %s", a.config.ServiceName, a.config.HTTPAddr)
		serverErr <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimout)
		defer cancel()

		err := a.server.Shutdown(shutdownCtx)
		if a.db != nil {
			_ = a.db.Close()
		}
		if a.redis != nil {
			_ = a.redis.Close()
		}

		return err
	case err := <-serverErr:
		if a.db != nil {
			_ = a.db.Close()
		}
		if a.redis != nil {
			_ = a.redis.Close()
		}

		if err == http.ErrServerClosed {
			return nil
		}

		return err
	}
}

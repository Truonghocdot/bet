package http

import (
	"net/http"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"

	"github.com/redis/go-redis/v9"
)

func NewRouter(
	_ any,
	authService *service.AuthService,
	affiliateService *service.AffiliateService,
	walletService *service.WalletService,
	notificationService *service.NotificationService,
	contentService *service.ContentService,
	sessionService *service.GameSessionService,
	betService *service.BetService,
	playRoomService *service.PlayRoomService,
	depositService *service.DepositService,
	withdrawalService *service.WithdrawalService,
	broker *realtime.Broker,
	gameRepository *repopg.GameRepository,
	redis *redis.Client,
	internalToken string,
) http.Handler {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler(authService)
	affiliateHandler := NewAffiliateHandler(affiliateService)
	walletHandler := NewWalletHandler(walletService, broker)
	notificationHandler := NewNotificationHandler(notificationService)
	contentHandler := NewContentHandler(contentService)
	gameHandler := NewGameHandler(sessionService, betService)
	playRoomHandler := NewPlayRoomHandler(playRoomService, broker)
	depositHandler := NewDepositHandler(depositService, internalToken)
	withdrawalHandler := NewWithdrawalHandler(withdrawalService)
	adminHandler := NewAdminHandler(gameRepository, broker, redis, authService)
	authSSOHandler := NewAuthSSOHandler(authService, redis)
	authn := authmiddleware.NewAuthentication(authService)

	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
	mux.HandleFunc("POST /v1/auth/reset-password", authHandler.ResetPassword)
	mux.Handle("GET /v1/auth/me", authn.Require(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /v1/affiliate/summary", authn.Require(http.HandlerFunc(affiliateHandler.Summary)))
	mux.Handle("POST /v1/affiliate/become-agency", authn.Require(http.HandlerFunc(affiliateHandler.BecomeAgency)))
	mux.Handle("GET /v1/wallets/summary", authn.Require(http.HandlerFunc(walletHandler.ServeHTTP)))
	mux.Handle("POST /v1/wallets/exchange", authn.Require(http.HandlerFunc(walletHandler.Exchange)))
	mux.Handle("GET /v1/wallets/stream", authn.Require(http.HandlerFunc(walletHandler.Stream)))
	mux.Handle("GET /v1/notifications", authn.Require(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("GET /v1/notifications/stream", authn.Require(http.HandlerFunc(notificationHandler.Stream)))
	mux.Handle("POST /v1/notifications/{id}/read", authn.Require(http.HandlerFunc(notificationHandler.MarkRead)))
	mux.HandleFunc("GET /v1/content/home", contentHandler.Home)
	mux.HandleFunc("GET /v1/content/promotions", contentHandler.Promotions)
	mux.HandleFunc("GET /v1/content/news", contentHandler.News)
	mux.HandleFunc("GET /v1/content/news/{slug}", contentHandler.NewsDetail)
	mux.HandleFunc("GET /v1/play/rooms", playRoomHandler.ListRooms)
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/state", playRoomHandler.RoomState)
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/stream", playRoomHandler.RoomStateStream)
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/ws", playRoomHandler.RoomStateWS)
	mux.Handle("GET /v1/play/rooms/{room_code}/bets/ws", authn.Require(http.HandlerFunc(playRoomHandler.MyBetsWS)))
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/history", playRoomHandler.RoomHistory)
	mux.Handle("GET /v1/play/rooms/{room_code}/bets", authn.Require(http.HandlerFunc(playRoomHandler.MyRoomBets)))
	mux.Handle("POST /v1/play/rooms/{room_code}/bets", authn.Require(http.HandlerFunc(playRoomHandler.PlaceRoomBet)))
	mux.Handle("GET /v1/games/", authn.Require(http.HandlerFunc(gameHandler.ServeHTTP)))
	mux.Handle("POST /v1/games/", authn.Require(http.HandlerFunc(gameHandler.ServeHTTP)))
	mux.Handle("POST /v1/deposits/", authn.Require(http.HandlerFunc(depositHandler.ServeHTTP)))
	mux.Handle("GET /v1/deposits/", authn.Require(http.HandlerFunc(depositHandler.ServeHTTP)))
	// Support both trailing-slash and non-trailing-slash variants for withdrawal routes.
	mux.Handle("POST /v1/withdrawals", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.Handle("GET /v1/withdrawals", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.Handle("DELETE /v1/withdrawals", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.Handle("POST /v1/withdrawals/", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.Handle("GET /v1/withdrawals/", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.Handle("DELETE /v1/withdrawals/", authn.Require(http.HandlerFunc(withdrawalHandler.ServeHTTP)))
	mux.HandleFunc("POST /internal/v1/deposits/apply", depositHandler.Apply)

	// Admin control routes
	requireAdmin := func(next http.Handler) http.Handler {
		return authn.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := authmiddleware.CurrentClaims(r.Context())
			if !ok || claims.Role != 1 {
				writeJSON(w, http.StatusForbidden, map[string]string{"message": "Quyền truy cập bị từ chối"})
				return
			}
			next.ServeHTTP(w, r)
		}))
	}

	mux.Handle("GET /v1/admin/rooms/stats", requireAdmin(http.HandlerFunc(adminHandler.ListRoomStats)))
	mux.Handle("GET /v1/admin/rooms/stats/stream", requireAdmin(http.HandlerFunc(adminHandler.StreamRoomStats)))
	mux.HandleFunc("GET /v1/admin/rooms/stats/ws", adminHandler.StreamRoomStatsWS)
	mux.Handle("POST /v1/admin/periods/{id}/result", requireAdmin(http.HandlerFunc(adminHandler.SetManualResult)))
	mux.Handle("POST /v1/admin/lock", requireAdmin(http.HandlerFunc(adminHandler.AcquireLock)))
	mux.Handle("PUT /v1/admin/lock", requireAdmin(http.HandlerFunc(adminHandler.HeartbeatLock)))
	mux.Handle("DELETE /v1/admin/lock", requireAdmin(http.HandlerFunc(adminHandler.ReleaseLock)))

	mux.HandleFunc("POST /v1/auth/sso/exchange", authSSOHandler.Exchange)

	return RecoverMiddleware(withCORS(mux))
}

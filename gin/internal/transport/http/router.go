package http

import (
	"net/http"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/service"
)

func NewRouter(
	_ any,
	authService *service.AuthService,
	walletService *service.WalletService,
	notificationService *service.NotificationService,
	sessionService *service.GameSessionService,
	betService *service.BetService,
	playRoomService *service.PlayRoomService,
	depositService *service.DepositService,
	internalToken string,
) http.Handler {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler(authService)
	walletHandler := NewWalletHandler(walletService)
	notificationHandler := NewNotificationHandler(notificationService)
	gameHandler := NewGameHandler(sessionService, betService)
	playRoomHandler := NewPlayRoomHandler(playRoomService)
	depositHandler := NewDepositHandler(depositService, internalToken)
	authn := authmiddleware.NewAuthentication(authService)

	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
	mux.HandleFunc("POST /v1/auth/reset-password", authHandler.ResetPassword)
	mux.Handle("GET /v1/auth/me", authn.Require(http.HandlerFunc(authHandler.Me)))
	mux.Handle("GET /v1/wallets/summary", authn.Require(http.HandlerFunc(walletHandler.ServeHTTP)))
	mux.Handle("GET /v1/notifications", authn.Require(http.HandlerFunc(notificationHandler.List)))
	mux.Handle("POST /v1/notifications/{id}/read", authn.Require(http.HandlerFunc(notificationHandler.MarkRead)))
	mux.HandleFunc("GET /v1/play/rooms", playRoomHandler.ListRooms)
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/state", playRoomHandler.RoomState)
	mux.HandleFunc("GET /v1/play/rooms/{room_code}/history", playRoomHandler.RoomHistory)
	mux.Handle("GET /v1/play/rooms/{room_code}/bets", authn.Require(http.HandlerFunc(playRoomHandler.MyRoomBets)))
	mux.Handle("POST /v1/play/rooms/{room_code}/bets", authn.Require(http.HandlerFunc(playRoomHandler.PlaceRoomBet)))
	mux.Handle("GET /v1/games/", authn.Require(http.HandlerFunc(gameHandler.ServeHTTP)))
	mux.Handle("POST /v1/games/", authn.Require(http.HandlerFunc(gameHandler.ServeHTTP)))
	mux.Handle("POST /v1/deposits/", authn.Require(http.HandlerFunc(depositHandler.ServeHTTP)))
	mux.HandleFunc("POST /internal/v1/deposits/apply", depositHandler.Apply)

	return withCORS(mux)
}

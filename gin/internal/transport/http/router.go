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
	sessionService *service.GameSessionService,
	betService *service.BetService,
	depositService *service.DepositService,
	internalToken string,
) http.Handler {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler()
	authHandler := NewAuthHandler(authService)
	walletHandler := NewWalletHandler(walletService)
	gameHandler := NewGameHandler(sessionService, betService)
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
	mux.Handle("POST /v1/games/", authn.Require(http.HandlerFunc(gameHandler.ServeHTTP)))
	mux.Handle("POST /v1/deposits/", authn.Require(http.HandlerFunc(depositHandler.ServeHTTP)))
	mux.HandleFunc("POST /internal/v1/deposits/apply", depositHandler.Apply)

	return withCORS(mux)
}

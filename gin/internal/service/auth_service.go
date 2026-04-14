package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gin/internal/auth/otp"
	"gin/internal/auth/password"
	"gin/internal/auth/token"
	"gin/internal/domain/auth"
	"gin/internal/integration/gate"
	repopg "gin/internal/repository/postgres"
	"gin/internal/security/ratelimit"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	"gin/internal/support/message"
)

var (
	ErrInvalidCredentials = errors.New(message.InvalidCredentials)
	ErrUnauthorized       = errors.New(message.Unauthorized)
	ErrRateLimited        = errors.New(message.TooManyRequests)
	ErrLoginLocked        = errors.New(message.LoginTemporarilyLocked)
	ErrOTPInvalid         = errors.New(message.OTPInvalid)
	ErrOTPExpired         = errors.New(message.OTPExpired)
	ErrOTPLocked          = errors.New(message.OTPLocked)
	ErrResetTokenInvalid  = errors.New(message.ResetTokenInvalid)
)

type AuthConfig struct {
	RegisterURL           string
	OTPSecret             string
	ForgotOTPTTL          time.Duration
	ForgotCooldown        time.Duration
	ForgotMaxAttempts     int
	ForgotWindow          time.Duration
	ForgotLimitIP         int
	ForgotLimitTarget     int
	LoginFailWindow       time.Duration
	LoginFailLimitIP      int
	LoginFailLimitAccount int
	LoginLockDuration     time.Duration
	RegisterWindow        time.Duration
	RegisterLimitIP       int
	RegisterLimitEmail    int
	RegisterLimitPhone    int
	RefreshTokenTTL       time.Duration
}

type AuthService struct {
	repository  *repopg.UserRepository
	tokenSigner *token.Signer
	limiter     *ratelimit.Limiter
	notifier    *gate.Notifier
	config      AuthConfig
}

func NewAuthService(
	repository *repopg.UserRepository,
	tokenSigner *token.Signer,
	limiter *ratelimit.Limiter,
	notifier *gate.Notifier,
	config AuthConfig,
) *AuthService {
	return &AuthService{
		repository:  repository,
		tokenSigner: tokenSigner,
		limiter:     limiter,
		notifier:    notifier,
		config:      config,
	}
}

func (s *AuthService) Register(ctx context.Context, request auth.RegisterRequest, meta auth.RequestMeta) (auth.AuthResponse, error) {
	name := strings.TrimSpace(request.Name)
	email := strings.ToLower(strings.TrimSpace(request.Email))
	passwordRaw := strings.TrimSpace(request.Password)

	if len(name) < 2 || len(name) > 100 {
		return auth.AuthResponse{}, fmt.Errorf(message.NameInvalid)
	}

	// Email bây giờ là nullable hoàn toàn, không cần tự sinh.
	if email != "" && !strings.Contains(email, "@") {
		return auth.AuthResponse{}, fmt.Errorf(message.EmailInvalid)
	}

	var phone *string
	if trimmedPhone := strings.TrimSpace(request.Phone); trimmedPhone != "" {
		if len(trimmedPhone) < 8 || len(trimmedPhone) > 20 {
			return auth.AuthResponse{}, fmt.Errorf(message.PhoneInvalid)
		}
		phone = &trimmedPhone
	} else {
		// Số điện thoại là bắt buộc nếu không có Email (hoặc bắt buộc luôn tùy spec)
		// Ở đây tôi coi phone là bắt buộc theo yêu cầu của bạn.
		return auth.AuthResponse{}, fmt.Errorf(message.PhoneRequired)
	}

	if len(passwordRaw) < 6 || len(passwordRaw) > 72 {
		return auth.AuthResponse{}, fmt.Errorf(message.PasswordInvalid)
	}

	if err := s.enforceRegisterRateLimit(ctx, meta, email, phone); err != nil {
		return auth.AuthResponse{}, err
	}

	hash, err := password.Hash(passwordRaw)
	if err != nil {
		return auth.AuthResponse{}, err
	}

	profile, err := s.repository.CreateRegisteredUser(ctx, repopg.RegisterUserParams{
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: hash,
		RefCode:      strings.TrimSpace(request.RefCode),
		RegisterURL:  s.config.RegisterURL,
	})
	if err != nil {
		return auth.AuthResponse{}, err
	}

	return s.newAuthResponse(ctx, profile, meta)
}

func (s *AuthService) Login(ctx context.Context, request auth.LoginRequest, meta auth.RequestMeta) (auth.AuthResponse, error) {
	account := strings.TrimSpace(request.Account)
	if account == "" {
		return auth.AuthResponse{}, fmt.Errorf(message.AccountRequired)
	}

	if strings.TrimSpace(request.Password) == "" {
		return auth.AuthResponse{}, fmt.Errorf(message.PasswordRequired)
	}

	if err := s.ensureLoginNotLocked(ctx, account); err != nil {
		return auth.AuthResponse{}, err
	}

	profile, hash, err := s.repository.FindByAccount(ctx, account)
	if err != nil {
		s.registerLoginFailure(ctx, meta, account)
		if errors.Is(err, repopg.ErrAccountNotFound) {
			return auth.AuthResponse{}, ErrInvalidCredentials
		}
		return auth.AuthResponse{}, err
	}

	if err := password.Compare(hash, request.Password); err != nil {
		s.registerLoginFailure(ctx, meta, account)
		return auth.AuthResponse{}, ErrInvalidCredentials
	}

	now := clock.Now()
	if err := s.repository.MarkLoggedIn(ctx, profile.User.ID, now); err != nil {
		return auth.AuthResponse{}, err
	}
	profile.User.LastLoginAt = &now
	_ = s.clearLoginFailureState(ctx, account, meta)

	return s.newAuthResponse(ctx, profile, meta)
}

func (s *AuthService) LoginByUserID(ctx context.Context, userID int64) (auth.AuthResponse, error) {
	profile, err := s.repository.FindProfileByUserID(ctx, userID)
	if err != nil {
		return auth.AuthResponse{}, err
	}

	now := clock.Now()
	if err := s.repository.MarkLoggedIn(ctx, profile.User.ID, now); err != nil {
		return auth.AuthResponse{}, err
	}
	profile.User.LastLoginAt = &now

	return s.newAuthResponse(ctx, profile, auth.RequestMeta{})
}


func (s *AuthService) ForgotPassword(ctx context.Context, request auth.ForgotPasswordRequest, meta auth.RequestMeta) (auth.MessageResponse, error) {
	channel, channelDBValue, err := s.parseChannel(request.Channel)
	if err != nil {
		return auth.MessageResponse{}, err
	}

	account := normalizeAccount(channel, request.Account)
	if account == "" {
		return auth.MessageResponse{}, fmt.Errorf(message.AccountRequired)
	}

	if err := s.enforceForgotPasswordRateLimit(ctx, meta, channel, account); err != nil {
		return auth.MessageResponse{}, err
	}

	profile, err := s.repository.FindUserByChannel(ctx, channel, account)
	if err != nil {
		if errors.Is(err, repopg.ErrAccountNotFound) {
			return auth.MessageResponse{Message: message.ForgotPasswordAccepted}, nil
		}

		return auth.MessageResponse{}, err
	}

	if profile.User.Status != 1 {
		return auth.MessageResponse{Message: message.ForgotPasswordAccepted}, nil
	}

	code, err := otp.GenerateCode(6)
	if err != nil {
		return auth.MessageResponse{}, err
	}

	requestToken, err := otp.NewRequestToken()
	if err != nil {
		return auth.MessageResponse{}, err
	}

	expiresAt := clock.Now().Add(s.config.ForgotOTPTTL)
	record, err := s.repository.CreateForgotPasswordOTP(
		ctx,
		profile.User.ID,
		channelDBValue,
		account,
		otp.Hash(s.config.OTPSecret, code),
		otp.Last4(code),
		requestToken,
		expiresAt,
		s.config.ForgotMaxAttempts,
	)
	if err != nil {
		return auth.MessageResponse{}, err
	}

	if err := s.dispatchForgotPasswordNotification(ctx, channel, account, profile.User.Name, code); err != nil {
		log.Printf("[auth] khong the gui otp reset password: %v", err)
		return auth.MessageResponse{Message: message.ForgotPasswordAccepted}, nil
	}

	_ = s.repository.MarkOTPSent(ctx, record.ID, clock.Now())

	return auth.MessageResponse{Message: message.ForgotPasswordAccepted}, nil
}

func (s *AuthService) VerifyForgotPasswordOTP(ctx context.Context, request auth.VerifyForgotPasswordOTPRequest) (auth.VerifyForgotPasswordOTPResponse, error) {
	channel, channelDBValue, err := s.parseChannel(request.Channel)
	if err != nil {
		return auth.VerifyForgotPasswordOTPResponse{}, err
	}

	account := normalizeAccount(channel, request.Account)
	if account == "" {
		return auth.VerifyForgotPasswordOTPResponse{}, fmt.Errorf(message.AccountRequired)
	}

	if strings.TrimSpace(request.OTP) == "" {
		return auth.VerifyForgotPasswordOTPResponse{}, fmt.Errorf(message.OTPRequired)
	}

	record, err := s.repository.FindLatestPendingOTP(ctx, channelDBValue, account)
	if err != nil {
		if errors.Is(err, repopg.ErrOTPNotFound) {
			return auth.VerifyForgotPasswordOTPResponse{}, ErrOTPInvalid
		}
		return auth.VerifyForgotPasswordOTPResponse{}, err
	}

	if clock.Now().After(record.ExpiresAt) {
		return auth.VerifyForgotPasswordOTPResponse{}, ErrOTPExpired
	}

	if record.Status == auth.OTPStatusLocked {
		return auth.VerifyForgotPasswordOTPResponse{}, ErrOTPLocked
	}

	if !otp.Compare(s.config.OTPSecret, record.OTPHash, strings.TrimSpace(request.OTP)) {
		record.AttemptCount++
		locked := record.AttemptCount >= record.MaxAttempts
		var lockedAt *time.Time
		if locked {
			now := clock.Now()
			lockedAt = &now
		}
		if err := s.repository.MarkOTPAttempt(ctx, record, locked, lockedAt); err != nil {
			return auth.VerifyForgotPasswordOTPResponse{}, err
		}
		if locked {
			return auth.VerifyForgotPasswordOTPResponse{}, ErrOTPLocked
		}
		return auth.VerifyForgotPasswordOTPResponse{}, ErrOTPInvalid
	}

	if err := s.repository.MarkOTPVerified(ctx, record.ID, clock.Now()); err != nil {
		return auth.VerifyForgotPasswordOTPResponse{}, err
	}

	return auth.VerifyForgotPasswordOTPResponse{
		Message:    message.OTPVerified,
		ResetToken: record.RequestToken,
		ExpiresIn:  int64(time.Until(record.ExpiresAt).Seconds()),
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, request auth.ResetPasswordRequest) (auth.MessageResponse, error) {
	if strings.TrimSpace(request.ResetToken) == "" {
		return auth.MessageResponse{}, fmt.Errorf(message.ResetTokenRequired)
	}

	passwordRaw := strings.TrimSpace(request.NewPassword)
	if len(passwordRaw) < 6 || len(passwordRaw) > 72 {
		return auth.MessageResponse{}, fmt.Errorf(message.PasswordInvalid)
	}

	hash, err := password.Hash(passwordRaw)
	if err != nil {
		return auth.MessageResponse{}, err
	}

	if err := s.repository.ResetPasswordWithVerifiedOTP(ctx, strings.TrimSpace(request.ResetToken), hash); err != nil {
		if errors.Is(err, repopg.ErrResetTokenInvalid) {
			return auth.MessageResponse{}, ErrResetTokenInvalid
		}
		if errors.Is(err, repopg.ErrOTPExpired) {
			return auth.MessageResponse{}, ErrOTPExpired
		}
		return auth.MessageResponse{}, err
	}

	return auth.MessageResponse{Message: message.ResetPasswordSuccess}, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (auth.UserProfile, error) {
	if userID == 0 {
		return auth.UserProfile{}, ErrUnauthorized
	}

	return s.repository.FindProfileByUserID(ctx, userID)
}

func (s *AuthService) VerifyAccessToken(tokenValue string) (auth.TokenClaims, error) {
	return s.tokenSigner.Verify(tokenValue)
}

func (s *AuthService) Refresh(ctx context.Context, request auth.RefreshTokenRequest, meta auth.RequestMeta) (auth.AuthResponse, error) {
	userID, expiresAt, err := s.repository.FindRefreshToken(ctx, request.RefreshToken)
	if err != nil {
		return auth.AuthResponse{}, err
	}

	if clock.Now().After(expiresAt) {
		_ = s.repository.DeleteRefreshToken(ctx, request.RefreshToken)
		return auth.AuthResponse{}, errors.New(message.Unauthorized)
	}

	profile, err := s.repository.FindProfileByUserID(ctx, userID)
	if err != nil {
		return auth.AuthResponse{}, err
	}

	// Rotate token: delete old one
	_ = s.repository.DeleteRefreshToken(ctx, request.RefreshToken)

	return s.newAuthResponse(ctx, profile, meta)
}

func (s *AuthService) enforceRegisterRateLimit(ctx context.Context, meta auth.RequestMeta, email string, phone *string) error {
	if strings.TrimSpace(meta.IP) != "" {
		result, err := s.limiter.HitWindow(ctx, "auth:rate:register:ip:"+meta.IP, int64(s.config.RegisterLimitIP), s.config.RegisterWindow)
		if err != nil {
			return err
		}
		if !result.Allowed {
			return ErrRateLimited
		}
	}

	result, err := s.limiter.HitWindow(ctx, "auth:rate:register:email:"+email, int64(s.config.RegisterLimitEmail), 24*time.Hour)
	if err != nil {
		return err
	}
	if !result.Allowed {
		return ErrRateLimited
	}

	if phone != nil {
		result, err := s.limiter.HitWindow(ctx, "auth:rate:register:phone:"+*phone, int64(s.config.RegisterLimitPhone), 24*time.Hour)
		if err != nil {
			return err
		}
		if !result.Allowed {
			return ErrRateLimited
		}
	}

	return nil
}

func (s *AuthService) ensureLoginNotLocked(ctx context.Context, account string) error {
	locked, _, err := s.limiter.IsLocked(ctx, "auth:lock:login:"+strings.ToLower(strings.TrimSpace(account)))
	if err != nil {
		return err
	}
	if locked {
		return ErrLoginLocked
	}
	return nil
}

func (s *AuthService) registerLoginFailure(ctx context.Context, meta auth.RequestMeta, account string) {
	if strings.TrimSpace(meta.IP) != "" {
		_, _ = s.limiter.HitWindow(ctx, "auth:rate:login:ip:"+meta.IP, int64(s.config.LoginFailLimitIP), s.config.LoginFailWindow)
	}

	accountKey := "auth:rate:login:account:" + strings.ToLower(strings.TrimSpace(account))
	result, err := s.limiter.HitWindow(ctx, accountKey, int64(s.config.LoginFailLimitAccount), s.config.LoginFailWindow)
	if err == nil && !result.Allowed {
		_ = s.limiter.Lock(ctx, "auth:lock:login:"+strings.ToLower(strings.TrimSpace(account)), s.config.LoginLockDuration)
	}
}

func (s *AuthService) clearLoginFailureState(ctx context.Context, account string, meta auth.RequestMeta) error {
	keys := []string{
		"auth:rate:login:account:" + strings.ToLower(strings.TrimSpace(account)),
		"auth:lock:login:" + strings.ToLower(strings.TrimSpace(account)),
	}
	if strings.TrimSpace(meta.IP) != "" {
		keys = append(keys, "auth:rate:login:ip:"+meta.IP)
	}
	return s.limiter.Clear(ctx, keys...)
}

func (s *AuthService) enforceForgotPasswordRateLimit(ctx context.Context, meta auth.RequestMeta, channel auth.OTPChannel, account string) error {
	if strings.TrimSpace(meta.IP) != "" {
		result, err := s.limiter.HitWindow(ctx, "auth:rate:forgot:ip:"+meta.IP, int64(s.config.ForgotLimitIP), s.config.ForgotWindow)
		if err != nil {
			return err
		}
		if !result.Allowed {
			return ErrRateLimited
		}
	}

	targetKey := "auth:rate:forgot:target:" + string(channel) + ":" + account
	result, err := s.limiter.HitWindow(ctx, targetKey, int64(s.config.ForgotLimitTarget), s.config.ForgotWindow)
	if err != nil {
		return err
	}
	if !result.Allowed {
		return ErrRateLimited
	}

	started, _, err := s.limiter.StartCooldown(ctx, "auth:cooldown:forgot:"+string(channel)+":"+account, s.config.ForgotCooldown)
	if err != nil {
		return err
	}
	if !started {
		return ErrRateLimited
	}

	return nil
}

func (s *AuthService) dispatchForgotPasswordNotification(ctx context.Context, channel auth.OTPChannel, account string, name string, code string) error {
	request := gate.NotificationRequest{
		Target: account,
		Meta: map[string]any{
			"otp":                code,
			"expired_in_seconds": int64(s.config.ForgotOTPTTL.Seconds()),
			"user_name":          name,
			"purpose":            "reset_password",
		},
	}

	switch channel {
	case auth.OTPChannelEmail:
		request.Channel = "email"
		request.Subject = "Mã OTP đặt lại mật khẩu"
		request.Message = fmt.Sprintf("Mã OTP đặt lại mật khẩu của bạn là %s. Mã có hiệu lực trong %d phút.", code, int(s.config.ForgotOTPTTL.Minutes()))
	case auth.OTPChannelPhone:
		request.Channel = "sms"
		request.Message = fmt.Sprintf("OTP dat lai mat khau cua ban la %s. Ma co hieu luc trong %d phut.", code, int(s.config.ForgotOTPTTL.Minutes()))
	default:
		return fmt.Errorf(message.OTPChannelInvalid)
	}

	return s.notifier.Send(ctx, request)
}

func (s *AuthService) parseChannel(channel auth.OTPChannel) (auth.OTPChannel, int, error) {
	switch auth.OTPChannel(strings.ToLower(strings.TrimSpace(string(channel)))) {
	case auth.OTPChannelEmail:
		return auth.OTPChannelEmail, 1, nil
	case auth.OTPChannelPhone:
		return auth.OTPChannelPhone, 2, nil
	default:
		return "", 0, fmt.Errorf(message.OTPChannelInvalid)
	}
}

func normalizeAccount(channel auth.OTPChannel, account string) string {
	trimmed := strings.TrimSpace(account)
	switch channel {
	case auth.OTPChannelEmail:
		return strings.ToLower(trimmed)
	case auth.OTPChannelPhone:
		return trimmed
	default:
		return trimmed
	}
}

func (s *AuthService) newAuthResponse(ctx context.Context, profile auth.UserProfile, meta auth.RequestMeta) (auth.AuthResponse, error) {
	issuedAt := clock.Now()
	expiresAt := issuedAt.Add(s.tokenSigner.TTL())

	accessToken, err := s.tokenSigner.Sign(auth.TokenClaims{
		UserID: profile.User.ID,
		Role:   profile.User.Role,
		Status: profile.User.Status,
		ExpAt:  expiresAt,
		Issued: issuedAt,
	})
	if err != nil {
		return auth.AuthResponse{}, err
	}

	refreshTokenValue := id.Long()
	refreshTTL := s.config.RefreshTokenTTL
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	refreshExpiresAt := issuedAt.Add(refreshTTL)

	err = s.repository.CreateRefreshToken(ctx, repopg.CreateRefreshTokenParams{
		UserID:    profile.User.ID,
		Token:     refreshTokenValue,
		ExpiresAt: refreshExpiresAt,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
	})
	if err != nil {
		return auth.AuthResponse{}, err
	}

	return auth.AuthResponse{
		User:             profile.User,
		Affiliate:        profile.AffiliateProfile,
		AccessToken:      accessToken,
		RefreshToken:     refreshTokenValue,
		TokenType:        "Bearer",
		ExpiresIn:        int64(time.Until(expiresAt).Seconds()),
		RefreshExpiresIn: int64(time.Until(refreshExpiresAt).Seconds()),
	}, nil
}

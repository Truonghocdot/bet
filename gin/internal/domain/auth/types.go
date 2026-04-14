package auth

import "time"

type OTPChannel string

const (
	OTPChannelEmail OTPChannel = "email"
	OTPChannelPhone OTPChannel = "phone"
)

const (
	OTPPurposeResetPassword = 1
)

const (
	OTPStatusPending   = 1
	OTPStatusVerified  = 2
	OTPStatusUsed      = 3
	OTPStatusExpired   = 4
	OTPStatusLocked    = 5
	OTPStatusCancelled = 6
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	RefCode  string `json:"ref_code"`
}

type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Channel OTPChannel `json:"channel"`
	Account string     `json:"account"`
}

type VerifyForgotPasswordOTPRequest struct {
	Channel OTPChannel `json:"channel"`
	Account string     `json:"account"`
	OTP     string     `json:"otp"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}

type RequestMeta struct {
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type VerifyForgotPasswordOTPResponse struct {
	Message    string `json:"message"`
	ResetToken string `json:"reset_token"`
	ExpiresIn  int64  `json:"expires_in"`
}

type User struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Phone           *string    `json:"phone,omitempty"`
	Role            int        `json:"role"`
	Status          int        `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AffiliateProfile struct {
	ID      int64  `json:"id"`
	RefCode string `json:"ref_code"`
	RefLink string `json:"ref_link"`
	Status  int    `json:"status"`
}

type UserProfile struct {
	User             User              `json:"user"`
	AffiliateProfile *AffiliateProfile `json:"affiliate_profile,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	User             User              `json:"user"`
	Affiliate        *AffiliateProfile `json:"affiliate_profile,omitempty"`
	AccessToken      string            `json:"access_token"`
	RefreshToken     string            `json:"refresh_token,omitempty"`
	TokenType        string            `json:"token_type"`
	ExpiresIn        int64             `json:"expires_in"`
	RefreshExpiresIn int64             `json:"refresh_expires_in,omitempty"`
}

type TokenClaims struct {
	UserID int64     `json:"user_id"`
	Role   int       `json:"role"`
	Status int       `json:"status"`
	ExpAt  time.Time `json:"exp_at"`
	Issued time.Time `json:"issued_at"`
}

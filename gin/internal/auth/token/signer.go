package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gin/internal/domain/auth"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type Signer struct {
	secret []byte
	ttl    time.Duration
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func NewSigner(secret string, ttl time.Duration) (*Signer, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf(message.AuthSecretRequired)
	}

	if ttl <= 0 {
		return nil, fmt.Errorf(message.AuthTTLRequired)
	}

	return &Signer{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (s *Signer) TTL() time.Duration {
	return s.ttl
}

func (s *Signer) Sign(claims auth.TokenClaims) (string, error) {
	headerBytes, err := json.Marshal(tokenHeader{
		Alg: "HS256",
		Typ: "JWT",
	})
	if err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	signature := s.sign(signingInput)

	return signingInput + "." + signature, nil
}

func (s *Signer) Verify(token string) (auth.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenFormatInvalid)
	}

	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.sign(signingInput))) {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenSignatureInvalid)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenPayloadInvalid)
	}

	var claims auth.TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenPayloadInvalid)
	}

	if claims.UserID == 0 {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenSubjectInvalid)
	}

	if clock.Now().After(claims.ExpAt) {
		return auth.TokenClaims{}, fmt.Errorf(message.TokenExpired)
	}

	return claims, nil
}

func (s *Signer) sign(input string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gin/internal/service"
	"github.com/redis/go-redis/v9"
)

type AuthSSOHandler struct {
	authService *service.AuthService
	redis       *redis.Client
}

func NewAuthSSOHandler(authService *service.AuthService, redis *redis.Client) *AuthSSOHandler {
	return &AuthSSOHandler{authService: authService, redis: redis}
}

func (h *AuthSSOHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Mã xác thực không hợp lệ"})
		return
	}

	ctx := r.Context()
	key := fmt.Sprintf("sso:token:%s", request.Token)
	
	val, err := h.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Mã xác thực đã hết hạn hoặc không tồn tại"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi hệ thống khi xác thực SSO"})
		return
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Dữ liệu người dùng không hợp lệ"})
		return
	}

	// Double check user and get JWT
	response, err := h.authService.LoginByUserID(ctx, userID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Không thể đăng nhập cho người dùng này"})
		return
	}

	// Burn the token after successful exchange
	h.redis.Del(ctx, key)

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthSSOHandler) CreateToken(ctx context.Context, userID int64) (string, error) {
	// This might be used if we wanted to generate SSO tokens FROM gin TO somewhere else.
	// But our current plan is Laravel -> Gin.
	return "", nil
}

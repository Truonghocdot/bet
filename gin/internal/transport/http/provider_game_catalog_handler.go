package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	"gin/internal/service"
	"gin/internal/support/message"
)

type ProviderGameCatalogHandler struct {
	service    *service.ProviderGameCatalogService
	tcgRuntime *service.TCGRuntimeService
}

func NewProviderGameCatalogHandler(service *service.ProviderGameCatalogService, tcgRuntime *service.TCGRuntimeService) *ProviderGameCatalogHandler {
	return &ProviderGameCatalogHandler{
		service:    service,
		tcgRuntime: tcgRuntime,
	}
}

func (h *ProviderGameCatalogHandler) TCGList(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": message.InternalServerError})
		return
	}

	productType := 0
	productTypeText := strings.TrimSpace(r.URL.Query().Get("product_type"))
	if productTypeText != "" {
		parsed, err := strconv.Atoi(productTypeText)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "product_type không hợp lệ"})
			return
		}
		productType = parsed
	}

	response, err := h.service.GetTCGGameCatalog(r.Context(), service.ProviderGameCatalogFilter{
		Category:        r.URL.Query().Get("category"),
		ProductType:     productType,
		IncludeChildren: isTruthyQueryValue(r.URL.Query().Get("include_children")),
	})
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func isTruthyQueryValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (h *ProviderGameCatalogHandler) Launch(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}
	if h.tcgRuntime == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": message.InternalServerError})
		return
	}

	var request game.ProviderGameLaunchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu launch game không hợp lệ"})
		return
	}

	response, err := h.tcgRuntime.LaunchGame(r.Context(), claims.UserID, service.TCGLaunchRequest{
		ProductType: request.ProductType,
		GameType:    request.GameType,
		GameCode:    request.GameCode,
		Name:        request.Name,
	}, game.ProviderGameLaunchMeta{
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, game.ProviderGameLaunchResponse{
		Message:           response.Message,
		GameURL:           response.GameURL,
		ProductType:       response.ProductType,
		GameType:          response.GameType,
		TransferredAmount: response.TransferredAmount,
		Source:            response.Source,
	})
}

func (h *ProviderGameCatalogHandler) CloseActive(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}
	if h.tcgRuntime == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": message.InternalServerError})
		return
	}

	response, err := h.tcgRuntime.CloseActive(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, game.ProviderGameCloseActiveResponse{
		Message:     response.Message,
		ProductType: response.ProductType,
		GameType:    response.GameType,
		Swept:       response.Swept,
		ReferenceNo: response.ReferenceNo,
	})
}

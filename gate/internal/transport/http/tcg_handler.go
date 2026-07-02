package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gate/internal/service"
)

type TCGHandler struct {
	tcgService    *service.TCGService
	internalToken string
}

func NewTCGHandler(tcgService *service.TCGService, internalToken string) *TCGHandler {
	return &TCGHandler{
		tcgService:    tcgService,
		internalToken: strings.TrimSpace(internalToken),
	}
}

func (h *TCGHandler) RegisterPlayer(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		log.Printf("[gate][tcg.register.error] reason=unauthorized path=%s", r.URL.Path)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	if h.tcgService == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "tcg service unavailable"})
		return
	}

	var request service.RegisterTCGPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[gate][tcg.register.error] reason=decode_failed path=%s err=%v", r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg register payload"})
		return
	}

	if err := h.tcgService.RegisterPlayer(r.Context(), request); err != nil {
		log.Printf("[gate][tcg.register.error] reason=register_failed username=%s err=%v", strings.TrimSpace(request.Username), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "tcg player registered"})
}

func (h *TCGHandler) Balance(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.TCGBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg balance payload"})
		return
	}

	response, err := h.tcgService.GetBalance(r.Context(), request)
	if err != nil && response.Status == 0 && response.ErrorDesc == "" {
		log.Printf("[gate][tcg.balance.error] username=%s err=%v", strings.TrimSpace(request.Username), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TCGHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.TCGTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg transfer payload"})
		return
	}

	response, err := h.tcgService.Transfer(r.Context(), request)
	if err != nil && response.Status == 0 && response.ErrorDesc == "" {
		log.Printf("[gate][tcg.transfer.error] username=%s ref=%s err=%v", strings.TrimSpace(request.Username), strings.TrimSpace(request.ReferenceNo), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TCGHandler) TransferOutAll(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.TCGTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg transfer out all payload"})
		return
	}

	response, err := h.tcgService.TransferOutAll(r.Context(), request)
	if err != nil && response.Status == 0 && response.ErrorDesc == "" {
		log.Printf("[gate][tcg.transfer_out_all.error] username=%s ref=%s err=%v", strings.TrimSpace(request.Username), strings.TrimSpace(request.ReferenceNo), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TCGHandler) TransferStatus(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.TCGTransferStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg transfer status payload"})
		return
	}

	response, err := h.tcgService.CheckTransferStatus(r.Context(), request)
	if err != nil && response.Status == 0 && response.ErrorDesc == "" && response.TransactionStatus == "" {
		log.Printf("[gate][tcg.transfer_status.error] ref=%s err=%v", strings.TrimSpace(request.ReferenceNo), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TCGHandler) LaunchGame(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.TCGLaunchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid tcg launch payload"})
		return
	}

	response, err := h.tcgService.LaunchGame(r.Context(), request)
	if err != nil && response.Status == 0 && response.ErrorDesc == "" && response.GameURL == "" {
		log.Printf("[gate][tcg.launch.error] username=%s product_type=%d game_code=%s err=%v", strings.TrimSpace(request.Username), request.ProductType, strings.TrimSpace(request.GameCode), err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TCGHandler) isAuthorized(token string) bool {
	expected := strings.TrimSpace(h.internalToken)
	if expected == "" {
		return false
	}

	return strings.TrimSpace(token) == expected
}

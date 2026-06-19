package http

import (
	"net/http"
	"strconv"
	"strings"

	"gin/internal/service"
	"gin/internal/support/message"
)

type FinanceFeedHandler struct {
	financeFeedService *service.FinanceFeedService
}

func NewFinanceFeedHandler(financeFeedService *service.FinanceFeedService) *FinanceFeedHandler {
	return &FinanceFeedHandler{financeFeedService: financeFeedService}
}

func (h *FinanceFeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channel := strings.TrimSpace(r.PathValue("channel"))
	if channel != "deposit" && channel != "withdraw" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
		return
	}

	limit := 12
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var (
		response any
		err      error
	)

	switch channel {
	case "deposit":
		response, err = h.financeFeedService.FakeDeposits(r.Context(), limit)
	case "withdraw":
		response, err = h.financeFeedService.FakeWithdraws(r.Context(), limit)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

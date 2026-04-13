package http

import (
	"net/http"
	"strconv"
	"strings"

	"gin/internal/service"
	"gin/internal/support/message"
)

type ContentHandler struct {
	contentService *service.ContentService
}

func NewContentHandler(contentService *service.ContentService) *ContentHandler {
	return &ContentHandler{contentService: contentService}
}

func (h *ContentHandler) Home(w http.ResponseWriter, r *http.Request) {
	response, err := h.contentService.Home(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) Promotions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := readContentPagination(r)
	response, err := h.contentService.Promotions(r.Context(), page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) News(w http.ResponseWriter, r *http.Request) {
	page, pageSize := readContentPagination(r)
	response, err := h.contentService.News(r.Context(), page, pageSize)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ContentHandler) NewsDetail(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.RouteNotFound})
		return
	}
	response, err := h.contentService.NewsDetail(r.Context(), slug)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func readContentPagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 10

	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			page = value
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("page_size")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			pageSize = value
		}
	}

	return page, pageSize
}

func (h *ContentHandler) writeError(w http.ResponseWriter, err error) {
	if service.IsContentNewsNotFound(err) {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.ContentNewsNotFound})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
}

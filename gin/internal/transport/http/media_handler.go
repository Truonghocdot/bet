package http

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type MediaHandler struct {
	popupVideoPath string
}

func NewMediaHandler(popupVideoPath string) *MediaHandler {
	return &MediaHandler{
		popupVideoPath: popupVideoPath,
	}
}

func (h *MediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/media/popup-video" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Route not found"})
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.handlePopupVideo(w, r)
}

func (h *MediaHandler) handlePopupVideo(w http.ResponseWriter, r *http.Request) {
	if h.popupVideoPath == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Popup video not configured"})
		return
	}

	file, err := os.Open(h.popupVideoPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Popup video not found"})
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil || info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "Popup video not found"})
		return
	}

	etag := popupVideoETag(info)
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	contentType := "video/mp4"
	if ext := filepath.Ext(info.Name()); ext == ".webm" {
		contentType = "video/webm"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Disposition", `inline; filename="popup-video`+filepath.Ext(info.Name())+`"`)

	http.ServeContent(w, r, info.Name(), info.ModTime(), file)
}

func popupVideoETag(info os.FileInfo) string {
	sum := sha1.Sum([]byte(info.Name() + "|" + info.ModTime().UTC().Format(time.RFC3339Nano) + "|" + fmt.Sprintf("%d", info.Size())))
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

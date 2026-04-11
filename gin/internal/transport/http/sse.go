package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type sseStream struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func newSSEStream(w http.ResponseWriter) (*sseStream, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming không được hỗ trợ")
	}

	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache, no-transform")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return &sseStream{
		w:       w,
		flusher: flusher,
	}, nil
}

func (s *sseStream) Event(name string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.EventRaw(name, data)
}

func (s *sseStream) EventRaw(name string, data []byte) error {
	if len(data) == 0 {
		data = []byte("null")
	}

	if _, err := fmt.Fprintf(s.w, "event: %s\n", name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

func (s *sseStream) KeepAlive() error {
	if _, err := fmt.Fprint(s.w, ": keep-alive\n\n"); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

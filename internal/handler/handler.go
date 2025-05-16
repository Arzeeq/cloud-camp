package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type dto struct {
	Token    string `json:"token"`
	Capacity int    `json:"capacity"`
}

type TokenServicer interface {
	SetCapacity(token string, capacity int) error
}

type TokenHandler struct {
	service TokenServicer
	l       *slog.Logger
}

func NewTokenHandler(service TokenServicer, l *slog.Logger) *TokenHandler {
	return &TokenHandler{service: service, l: l}
}

func (h *TokenHandler) SetCapacity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	var d dto
	if err := parse(r.Body, &d); err != nil {
		w.Write([]byte("failed to parse request parameters"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := h.service.SetCapacity(d.Token, d.Capacity)
	if err != nil {
		h.l.Error(fmt.Sprintf("Failed to set capacity: %v", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parse(r io.Reader, payload any) error {
	if r == nil {
		return fmt.Errorf("parsing from nil reader")
	}

	err := json.NewDecoder(r).Decode(payload)
	if err != nil {
		return fmt.Errorf("failed to decode json: %v", err)
	}

	return nil
}

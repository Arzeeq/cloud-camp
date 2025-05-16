package ratelimiter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Bucketer interface {
	Take(token string) bool
}

type RateLimiter struct {
	b Bucketer
	l *slog.Logger
}

func New(b Bucketer, l *slog.Logger) (*RateLimiter, error) {
	return &RateLimiter{b: b, l: l}, nil
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")

		if key == "" {
			rl.writeResponse(w, response{
				Code:    http.StatusBadRequest,
				Message: "No X-API-Key provided",
			})
			return
		}

		if rl.b.Take(key) {
			next.ServeHTTP(w, r)
			return
		}

		rl.writeResponse(w, response{
			Code:    http.StatusTooManyRequests,
			Message: "Rate limit exceeded",
		})
	})
}

func (rl *RateLimiter) writeResponse(w http.ResponseWriter, res response) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(res.Code)

	err := json.NewEncoder(w).Encode(res)

	if err != nil {
		rl.l.Error(fmt.Sprintf("failed to write response: %v", err))
	}
}

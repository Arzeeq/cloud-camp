package main

import (
	"net/http"
	"time"

	"github.com/Arzeeq/cloud-camp/internal/bucket"
	"github.com/Arzeeq/cloud-camp/internal/logger"
	"github.com/Arzeeq/cloud-camp/internal/ratelimiter"
)

type handler int

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func main() {
	l := logger.New("text", "info")

	b := bucket.New(2, 60*time.Second)
	defer b.Stop()
	limiter, err := ratelimiter.New(b, l)
	if err != nil {
		panic(err)
	}

	h := handler(0)

	http.ListenAndServe(":8080", limiter.Middleware(&h))
}

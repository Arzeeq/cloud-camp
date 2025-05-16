package loadbalancer

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Pooler interface {
	Get() (string, error)
}

type LoadBalancer struct {
	pool Pooler
	l    *slog.Logger
}

func New(pool Pooler, logger *slog.Logger) (*LoadBalancer, error) {
	if pool == nil || logger == nil {
		return nil, errors.New("nil values in Load Balancer constructor")
	}

	return &LoadBalancer{
		pool: pool,
		l:    logger,
	}, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server, err := lb.pool.Get()
	if err != nil {
		lb.l.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		lb.l.Error("failed to parse host's url received from pool", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.NewSingleHostReverseProxy(serverURL).ServeHTTP(w, r)
}

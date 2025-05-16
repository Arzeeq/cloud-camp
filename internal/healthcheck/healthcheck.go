package healthcheck

import (
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Pooler interface {
	// GetAll return URLs of all alive and dead servers
	GetAll() []string
	// Enable returns true if server were marked as dead before
	Enable(string) bool
	// Disable returns true if server were marked as alive before
	Disable(string) bool
}

type HealthCheck struct {
	pool     Pooler
	interval time.Duration
	done     chan struct{}
	l        *slog.Logger
}

func New(pool Pooler, logger *slog.Logger, interval time.Duration) *HealthCheck {
	return &HealthCheck{
		pool:     pool,
		interval: interval,
		done:     make(chan struct{}),
		l:        logger,
	}
}

func (hc *HealthCheck) Start() {
	go hc.start()
}

func (hc *HealthCheck) Stop() {
	close(hc.done)
}

func (hc *HealthCheck) start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	hc.checkAll()
	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-hc.done:
			return
		}
	}
}

func (hc *HealthCheck) checkAll() {
	var wg sync.WaitGroup

	for _, s := range hc.pool.GetAll() {
		wg.Add(1)
		go func() {
			isAlive := hc.check(s)
			if isAlive && hc.pool.Enable(s) {
				hc.l.Info("Server is alive", slog.String("URL", s))
			} else if !isAlive && hc.pool.Disable(s) {
				hc.l.Info("Server is dead", slog.String("URL", s))
			}

			wg.Done()
		}()
	}
	wg.Wait()
}

func (hc *HealthCheck) check(url string) bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode >= 500 {
		return false
	}

	return true
}

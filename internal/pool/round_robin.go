package pool

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
)

type RoundRobinPool struct {
	mu         sync.RWMutex
	urls       []string
	healthy    map[string]bool
	idx        int
	aliveCount int
}

func NewRoundRobinPool(servers []string) (*RoundRobinPool, error) {
	urls := make([]string, len(servers))
	healthy := make(map[string]bool, len(servers))

	for i, s := range servers {
		_, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse server '%s': %v", s, err)
		}

		urls[i] = s
		healthy[s] = true
	}

	return &RoundRobinPool{
		urls:       urls,
		healthy:    healthy,
		aliveCount: len(urls),
	}, nil
}

func (p *RoundRobinPool) Get() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.idx >= p.aliveCount {
		p.idx = 0
	}

	aliveIdx := 0
	for i := range p.urls {
		if p.healthy[p.urls[i]] && aliveIdx == p.idx {
			p.idx++

			return p.urls[i], nil
		}

		if p.healthy[p.urls[i]] {
			aliveIdx++
		}
	}

	return "", errors.New("no server was found")
}

// GetAll return URLs of all alive and dead servers
func (p *RoundRobinPool) GetAll() []string {
	res := make([]string, len(p.urls))

	p.mu.RLock()
	defer p.mu.RUnlock()

	copy(res, p.urls)

	return res
}

// Enable returns true if server were marked as dead before
func (p *RoundRobinPool) Enable(server string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	prev, ok := p.healthy[server]
	if !ok {
		return false
	}

	if !prev {
		p.aliveCount++
	}
	p.healthy[server] = true

	return !prev
}

// Disable returns true if server were marked as alive before
func (p *RoundRobinPool) Disable(server string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	prev, ok := p.healthy[server]
	if !ok {
		return false
	}

	if prev {
		p.aliveCount--
	}
	p.healthy[server] = false

	return prev
}

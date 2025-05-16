package bucket

import (
	"sync"
	"time"
)

type TokenServicer interface {
	GetCapacity(token string) (int, error)
}

type Bucket struct {
	defaultCapacity int
	tokens          map[string]int
	lastAccess      map[string]time.Time
	mutex           sync.Mutex
	ticker          *time.Ticker
	done            chan struct{}
	tokenService    TokenServicer
}

func New(defaultCapacity int, interval time.Duration, tokenService TokenServicer) *Bucket {
	b := &Bucket{
		defaultCapacity: defaultCapacity,
		tokens:          make(map[string]int),
		lastAccess:      make(map[string]time.Time),
		done:            make(chan struct{}),
		tokenService:    tokenService,
	}

	b.ticker = time.NewTicker(interval)
	go b.refill()

	return b
}

func (b *Bucket) refill() {
	for {
		select {
		case <-b.ticker.C:
			b.mutex.Lock()

			for key := range b.tokens {
				if time.Since(b.lastAccess[key]) > time.Hour {
					delete(b.tokens, key)
					delete(b.lastAccess, key)
					continue
				}

				capacity, err := b.tokenService.GetCapacity(key)
				if err != nil {
					b.tokens[key] = b.defaultCapacity
				} else {
					b.tokens[key] = capacity
				}
			}

			b.mutex.Unlock()
		case <-b.done:
			return
		}
	}
}

func (b *Bucket) Take(token string) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.tokens[token]; !ok {
		capacity, err := b.tokenService.GetCapacity(token)
		if err != nil {
			b.tokens[token] = b.defaultCapacity
		} else {
			b.tokens[token] = capacity
		}
	}

	if b.tokens[token] > 0 {
		b.tokens[token]--
		b.lastAccess[token] = time.Now()
		return true
	}
	return false
}

func (b *Bucket) Stop() {
	b.ticker.Stop()
	close(b.done)
}

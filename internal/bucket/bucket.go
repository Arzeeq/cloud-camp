package bucket

import (
	"sync"
	"time"
)

type Bucket struct {
	capacity   int
	tokens     map[string]int
	lastAccess map[string]time.Time
	mutex      sync.Mutex
	ticker     *time.Ticker
	done       chan struct{}
}

func New(capacity int, interval time.Duration) *Bucket {
	b := &Bucket{
		capacity:   capacity,
		tokens:     make(map[string]int),
		lastAccess: make(map[string]time.Time),
		done:       make(chan struct{}),
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
				} else {
					b.tokens[key] = b.capacity
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
		b.tokens[token] = b.capacity
	}

	if b.tokens[token] > 0 {
		b.tokens[token]--
		return true
	}
	return false
}

func (b *Bucket) Stop() {
	b.ticker.Stop()
	close(b.done)
}

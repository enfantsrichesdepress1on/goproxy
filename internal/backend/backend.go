package backend

import (
	"errors"
	"math/rand"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Strategy int

const (
	StrategyRoundRobin Strategy = iota
	StrategyRandom
	StrategyLeastConnections
)

func ParseStrategy(s string) Strategy {
	switch s {
	case "random":
		return StrategyRandom
	case "least_connections":
		return StrategyLeastConnections
	default:
		return StrategyRoundRobin
	}
}

type Backend struct {
	URL *url.URL

	alive       int32
	activeConns int64
	totalReq    uint64
	totalErr    uint64
}

func NewBackend(rawURL string) (*Backend, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Backend{
		URL:   u,
		alive: 1,
	}, nil
}

func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

func (b *Backend) SetAlive(v bool) {
	if v {
		atomic.StoreInt32(&b.alive, 1)
	} else {
		atomic.StoreInt32(&b.alive, 0)
	}
}

func (b *Backend) IncActive() { atomic.AddInt64(&b.activeConns, 1) }
func (b *Backend) DecActive() { atomic.AddInt64(&b.activeConns, -1) }
func (b *Backend) ActiveConns() int64 {
	return atomic.LoadInt64(&b.activeConns)
}
func (b *Backend) IncReq() { atomic.AddUint64(&b.totalReq, 1) }
func (b *Backend) IncErr() { atomic.AddUint64(&b.totalErr, 1) }

type Status struct {
	URL           string `json:"url"`
	Alive         bool   `json:"alive"`
	ActiveConns   int64  `json:"active_conns"`
	TotalRequests uint64 `json:"total_requests"`
	TotalErrors   uint64 `json:"total_errors"`
}

func (b *Backend) Status() Status {
	return Status{
		URL:           b.URL.String(),
		Alive:         b.IsAlive(),
		ActiveConns:   b.ActiveConns(),
		TotalRequests: atomic.LoadUint64(&b.totalReq),
		TotalErrors:   atomic.LoadUint64(&b.totalErr),
	}
}

type Pool struct {
	mu       sync.RWMutex
	backends []*Backend
	strategy Strategy

	rrIndex uint64
}

func NewPool(urls []string, strategy Strategy) (*Pool, error) {
	if len(urls) == 0 {
		return nil, errors.New("no backends")
	}
	bs := make([]*Backend, 0, len(urls))
	for _, u := range urls {
		b, err := NewBackend(u)
		if err != nil {
			return nil, err
		}
		bs = append(bs, b)
	}
	rand.Seed(time.Now().UnixNano())
	return &Pool{
		backends: bs,
		strategy: strategy,
	}, nil
}

func (p *Pool) Backends() []*Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Backend, len(p.backends))
	copy(out, p.backends)
	return out
}

func (p *Pool) Statuses() []Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Status, 0, len(p.backends))
	for _, b := range p.backends {
		out = append(out, b.Status())
	}
	return out
}

func (p *Pool) NextBackend() *Backend {
	p.mu.RLock()
	defer p.mu.RUnlock()

	switch p.strategy {
	case StrategyRandom:
		return p.nextRandom()
	case StrategyLeastConnections:
		return p.nextLeastConnections()
	default:
		return p.nextRoundRobin()
	}
}

func (p *Pool) nextRoundRobin() *Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		idx := atomic.AddUint64(&p.rrIndex, 1)
		b := p.backends[int(idx-1)%n]
		if b.IsAlive() {
			return b
		}
	}
	return nil
}

func (p *Pool) nextRandom() *Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}
	alive := make([]*Backend, 0, n)
	for _, b := range p.backends {
		if b.IsAlive() {
			alive = append(alive, b)
		}
	}
	if len(alive) == 0 {
		return nil
	}
	return alive[rand.Intn(len(alive))]
}

func (p *Pool) nextLeastConnections() *Backend {
	n := len(p.backends)
	if n == 0 {
		return nil
	}
	var best *Backend
	var bestConns int64
	for _, b := range p.backends {
		if !b.IsAlive() {
			continue
		}
		conns := b.ActiveConns()
		if best == nil || conns < bestConns {
			best = b
			bestConns = conns
		}
	}
	return best
}

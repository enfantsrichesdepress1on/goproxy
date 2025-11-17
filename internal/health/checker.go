package health

import (
	"context"
	"net/http"
	"time"

	"github.com/enfantsrichesdepress1on/goproxy/internal/backend"
	"github.com/enfantsrichesdepress1on/goproxy/internal/logger"
)

type Checker struct {
	pool     *backend.Pool
	log      *logger.Logger
	client   *http.Client
	path     string
	interval time.Duration
}

func NewChecker(pool *backend.Pool, log *logger.Logger, path string, interval, timeout time.Duration) *Checker {
	return &Checker{
		pool: pool,
		log:  log,
		client: &http.Client{
			Timeout: timeout,
		},
		path:     path,
		interval: interval,
	}
}

func (c *Checker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		c.log.Infof("health checker started, interval=%s", c.interval)

		for {
			select {
			case <-ctx.Done():
				c.log.Infof("health checker stopped")
				return
			case <-ticker.C:
				c.checkAllOnce()
			}
		}
	}()
}

func (c *Checker) checkAllOnce() {
	backends := c.pool.Backends()
	for _, b := range backends {
		c.checkOne(b)
	}
}

func (c *Checker) checkOne(b *backend.Backend) {
	u := *b.URL
	u.Path = c.path

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		c.log.Errorf("health check build request error backend=%s err=%v", b.URL.String(), err)
		return
	}

	resp, err := c.client.Do(req)
	isUp := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 400
	if resp != nil {
		resp.Body.Close()
	}

	prevAlive := b.IsAlive()
	b.SetAlive(isUp)

	if isUp && !prevAlive {
		c.log.Infof("backend UP: %s", b.URL.String())
	}
	if !isUp && prevAlive {
		c.log.Errorf("backend DOWN: %s err=%v", b.URL.String(), err)
	}
}

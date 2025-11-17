package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/enfantsrichesdepress1on/goproxy/internal/api"
	"github.com/enfantsrichesdepress1on/goproxy/internal/backend"
	"github.com/enfantsrichesdepress1on/goproxy/internal/config"
	"github.com/enfantsrichesdepress1on/goproxy/internal/health"
	"github.com/enfantsrichesdepress1on/goproxy/internal/logger"
)

func main() {
	cfgPath := "config.json"
	if v := os.Getenv("GOPROXY_CONFIG"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic(err)
	}

	log := logger.New(cfg.LogLevel)

	log.Infof("config loaded listen_addr=%s strategy=%s backends=%d",
		cfg.ListenAddr, cfg.LoadBalancing.Strategy, len(cfg.Backends))

	backendURLs := make([]string, 0, len(cfg.Backends))
	for _, b := range cfg.Backends {
		backendURLs = append(backendURLs, b.URL)
	}

	pool, err := backend.NewPool(backendURLs, backend.ParseStrategy(cfg.LoadBalancing.Strategy))
	if err != nil {
		log.Errorf("failed to create backend pool: %v", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	healthInterval := time.Duration(cfg.HealthCheck.IntervalMs) * time.Millisecond
	healthTimeout := time.Duration(cfg.HealthCheck.TimeoutMs) * time.Millisecond
	requestTimeout := time.Duration(cfg.RequestTimeoutMs) * time.Millisecond

	checker := health.NewChecker(pool, log, cfg.HealthCheck.Path, healthInterval, healthTimeout)
	checker.Start(ctx)

	server := api.NewServer(pool, log, requestTimeout)

	httpServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: server,
	}

	go func() {
		<-ctx.Done()
		log.Infof("shutting down http server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Errorf("http server shutdown error: %v", err)
		}
	}()

	log.Infof("goproxy listening on %s", cfg.ListenAddr)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Errorf("http server error: %v", err)
	}
}

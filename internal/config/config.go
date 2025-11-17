package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type LoadBalancingConfig struct {
	Strategy string `json:"strategy"` // round_robin | random | least_connections
}

type BackendConfig struct {
	URL string `json:"url"`
}

type HealthCheckConfig struct {
	Path       string `json:"path"`        // /health
	IntervalMs int    `json:"interval_ms"` // 2000
	TimeoutMs  int    `json:"timeout_ms"`  // 1000
}

type Config struct {
	ListenAddr       string              `json:"listen_addr"`        // ":8080"
	LogLevel         string              `json:"log_level"`          // "debug"|"info"|"error"
	RequestTimeoutMs int                 `json:"request_timeout_ms"` // timeout для запросов к backend
	LoadBalancing    LoadBalancingConfig `json:"load_balancing"`
	Backends         []BackendConfig     `json:"backends"`
	HealthCheck      HealthCheckConfig   `json:"health_check"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.RequestTimeoutMs <= 0 {
		cfg.RequestTimeoutMs = 5000
	}
	if cfg.HealthCheck.IntervalMs <= 0 {
		cfg.HealthCheck.IntervalMs = 2000
	}
	if cfg.HealthCheck.TimeoutMs <= 0 {
		cfg.HealthCheck.TimeoutMs = 1000
	}
	if cfg.HealthCheck.Path == "" {
		cfg.HealthCheck.Path = "/health"
	}
	if cfg.LoadBalancing.Strategy == "" {
		cfg.LoadBalancing.Strategy = "round_robin"
	}
	if len(cfg.Backends) == 0 {
		return nil, fmt.Errorf("no backends configured")
	}

	return &cfg, nil
}

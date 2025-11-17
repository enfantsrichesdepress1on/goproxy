package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/enfantsrichesdepress1on/goproxy/internal/backend"
	"github.com/enfantsrichesdepress1on/goproxy/internal/logger"
)

type ProxyHandler struct {
	pool   *backend.Pool
	log    *logger.Logger
	client *http.Client
}

func NewProxyHandler(pool *backend.Pool, log *logger.Logger, timeout time.Duration) *ProxyHandler {
	return &ProxyHandler{
		pool: pool,
		log:  log,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	b := h.pool.NextBackend()
	if b == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}

	b.IncReq()
	b.IncActive()
	defer b.DecActive()

	outReq := r.Clone(r.Context())
	targetURL := *r.URL
	targetURL.Scheme = b.URL.Scheme
	targetURL.Host = b.URL.Host
	outReq.URL = &targetURL
	outReq.RequestURI = ""

	outReq.Header.Add("X-Forwarded-For", r.RemoteAddr)

	resp, err := h.client.Do(outReq)
	if err != nil {
		b.IncErr()
		h.log.Errorf("proxy error backend=%s method=%s path=%s err=%v",
			b.URL.String(), r.Method, r.URL.Path, err)
		http.Error(w, "backend error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		h.log.Debugf("failed to copy response body: %v", err)
	}

	duration := time.Since(start)
	h.log.Infof("proxy %s %s -> %s status=%d duration=%s",
		r.Method, r.URL.Path, b.URL.String(), resp.StatusCode, duration)
}

type Server struct {
	mux   *http.ServeMux
	pool  *backend.Pool
	log   *logger.Logger
	proxy *ProxyHandler
}

func NewServer(pool *backend.Pool, log *logger.Logger, proxyTimeout time.Duration) *Server {
	mux := http.NewServeMux()
	proxy := NewProxyHandler(pool, log, proxyTimeout)

	s := &Server{
		mux:   mux,
		pool:  pool,
		log:   log,
		proxy: proxy,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/backends", s.handleBackends)
	s.mux.Handle("/", s.proxy)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleBackends(w http.ResponseWriter, r *http.Request) {
	statuses := s.pool.Statuses()
	writeJSON(w, http.StatusOK, statuses)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

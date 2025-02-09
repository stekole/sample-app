package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type server struct {
	metrics *metrics
	srv     *http.Server
	logger  zerolog.Logger
}

type responseWriter struct {
	http.ResponseWriter
	status int
}
type metrics struct {
	httpRequestsTotal *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"endpoint", "method", "code"},
		),
		httpDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "method"},
		),
	}
	reg.MustRegister(m.httpRequestsTotal, m.httpDuration)
	return m
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func newServer(addr string, m *metrics, logger zerolog.Logger) *server {
	mux := http.NewServeMux()
	s := &server{
		metrics: m,
		srv: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		logger: logger,
	}
	mux.HandleFunc("/hello", s.helloHandler)
	mux.Handle("/metrics", promhttp.Handler())
	return s
}

func (s *server) helloHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := uuid.New().String()
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	defer func() {
		s.logger.Info().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status", rw.status).
			Float64("latency_ms", float64(time.Since(start).Milliseconds())).
			Msg("request completed")
	}()

	timer := prometheus.NewTimer(s.metrics.httpDuration.WithLabelValues("/hello", r.Method))
	defer timer.ObserveDuration()

	if r.Method != http.MethodGet {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		s.metrics.httpRequestsTotal.WithLabelValues("/hello", r.Method, "405").Inc()
		return
	}

	if r.Header.Get("Fail") == "true" {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		s.metrics.httpRequestsTotal.WithLabelValues("/hello", r.Method, "500").Inc()
		return
	}

	s.metrics.httpRequestsTotal.WithLabelValues("/hello", r.Method, "200").Inc()
	fmt.Fprintln(rw, "Hello, World!")
}

func (s *server) run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Msgf("Server shutdown error: %v", err)
		}
	}()

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newMetrics(prometheus.DefaultRegisterer)
	srv := newServer(":8080", m, logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	logger.Info().Msg("Server starting on :8080")
	if err := srv.run(ctx); err != nil {
		logger.Fatal().Msg(err.Error())
	}
}

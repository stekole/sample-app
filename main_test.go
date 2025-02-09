package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		headers      map[string]string
		wantStatus   int
		wantBody     string
		wantContains string
	}{
		{
			name:       "successful GET request",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "Hello, World!\n",
		},
		{
			name:         "method not allowed",
			method:       http.MethodPost,
			wantStatus:   http.StatusMethodNotAllowed,
			wantContains: "Method not allowed",
		},
		{
			name:         "forced failure",
			method:       http.MethodGet,
			headers:      map[string]string{"Fail": "true"},
			wantStatus:   http.StatusInternalServerError,
			wantContains: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			metrics := newMetrics(reg)
			logger := zerolog.New(io.Discard)
			srv := newServer(":8080", metrics, logger)

			req := httptest.NewRequest(tt.method, "/hello", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			srv.helloHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("handler returned wrong body: got %v want %v", w.Body.String(), tt.wantBody)
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("handler body doesn't contain expected string: got %v want to contain %v", w.Body.String(), tt.wantContains)
			}
		})
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	reg := prometheus.NewRegistry()
	metrics := newMetrics(reg)
	logger := zerolog.New(io.Discard)
	srv := newServer(":0", metrics, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error)
	go func() {
		errCh <- srv.run(ctx)
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("server.run() returned unexpected error: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Error("server didn't shut down within expected timeframe")
	}
}

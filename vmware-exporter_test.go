package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prezhdarov/prometheus-exporter/pkg/exporter"
)

func TestWebConfigUsesListenAddress(t *testing.T) {
	addr := ":9999"

	cfg := webConfig(&addr)

	if cfg == nil {
		t.Fatal("webConfig() returned nil")
	}

	if cfg.WebListenAddresses == nil {
		t.Fatal("WebListenAddresses is nil")
	}

	if len(*cfg.WebListenAddresses) != 1 {
		t.Fatalf("expected exactly 1 listen address, got %d", len(*cfg.WebListenAddresses))
	}

	if got := (*cfg.WebListenAddresses)[0]; got != addr {
		t.Fatalf("listen address = %q, want %q", got, addr)
	}

	if cfg.WebSystemdSocket == nil {
		t.Fatal("WebSystemdSocket is nil")
	}

	if *cfg.WebSystemdSocket {
		t.Fatal("WebSystemdSocket = true, want false")
	}

	if cfg.WebConfigFile == nil {
		t.Fatal("WebConfigFile is nil")
	}

	if *cfg.WebConfigFile != "" {
		t.Fatalf("WebConfigFile = %q, want empty string", *cfg.WebConfigFile)
	}
}

func TestProbeHandlerReturnsBadRequestWithoutTarget(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()

	exporter.CreateHandleFunc(rec, req, namespace, "", logger)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	if !strings.Contains(rec.Body.String(), "target parameter is required") {
		t.Fatalf("response body = %q, want it to contain %q", rec.Body.String(), "target parameter is required")
	}
}

func TestProbeHandlerWithTargetReturnsMetricsPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	req := httptest.NewRequest(http.MethodGet, "/probe?target=127.0.0.1:1", nil)
	rec := httptest.NewRecorder()

	exporter.CreateHandleFunc(rec, req, namespace, "", logger)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	if !strings.Contains(rec.Body.String(), "vmware_exporter_build_info") {
		t.Fatalf("response body = %q, want it to contain %q", rec.Body.String(), "vmware_exporter_build_info")
	}
}

package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/ksysoev/wasabi/tests"
)

func TestNewTracedMiddleware_WithoutExporter(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	exporter := func(_ context.Context) (sdktrace.SpanExporter, error) {
		return nil, errors.New("no exporter provided")
	}

	middleware := NewTracedMiddleware(exporter)
	handler := middleware(nextHandler)

	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	w1 := httptest.NewRecorder()

	tests.AssertPanic(t, func() { handler.ServeHTTP(w1, req1) }, "failed to initialize exporter: no exporter provided")
}

func TestNewTracedMiddleware_WithExporter(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	exporter := func(_ context.Context) (sdktrace.SpanExporter, error) {
		return stdouttrace.New()
	}

	middleware := NewTracedMiddleware(exporter)
	handler := middleware(nextHandler)

	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)

	resp1 := w1.Result()

	defer resp1.Body.Close()

	// assert request is successul
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	if Tracer == nil {
		t.Errorf("Tracer was not initialized")
	}
}

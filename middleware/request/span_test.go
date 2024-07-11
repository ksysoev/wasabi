package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksysoev/wasabi/tests"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestNewSpanMiddleware_TracerNotInitialized(t *testing.T) {
	spanName := "test-span"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewSpanMiddleware(spanName, nil)
	handler := middleware(nextHandler)

	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	w1 := httptest.NewRecorder()

	tests.AssertPanic(t, func() { handler.ServeHTTP(w1, req1) }, "NewSpanMiddleware called without initializing Tracer! Have you specified `WithTracer` server option?")
}

func TestNewSpanMiddleware_WithTracer(t *testing.T) {
	spanName := "test-span"
	exp, _ := stdouttrace.New()
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	tracer := tp.Tracer("test-tracer")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := NewSpanMiddleware(spanName, tracer)
	handler := middleware(nextHandler)

	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)

	resp1 := w1.Result()

	span := trace.SpanFromContext(req1.Context())

	defer resp1.Body.Close()

	// assert request is successul
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// assert span attributes
	if span.IsRecording() != false {
		t.Errorf("Since span has ended it should not record anymore")
	}
}

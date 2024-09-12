package request

import (
	"context"
	"fmt"
	"testing"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSpanMiddleware_TracerNotInitialized(t *testing.T) {
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return nil
	})

	spanName := "test-span"

	middleware := NewSpanMiddleware(spanName, nil)
	handler := middleware(mockHandler)

	panicFunction := func() {
		err := handler.Handle(mockConn, mockReq)
		if err != nil {
			fmt.Printf("Got error: %+v", err)
		}
	}

	assert.Panicsf(t, panicFunction, "NewSpanMiddleware called without initializing Tracer! Have you specified `WithTracer` server option?")
}

func TestNewSpanMiddleware_WithTracer(t *testing.T) {
	ctx := context.Background()
	spanName := "test-span"
	exp, _ := stdouttrace.New()
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	tracer := tp.Tracer("test-tracer")

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx)
	mockReq.EXPECT().WithContext(mock.AnythingOfType("*context.valueCtx")).Return(mockReq)

	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return nil
	})

	middleware := NewSpanMiddleware(spanName, tracer)
	handler := middleware(mockHandler)

	span := trace.SpanFromContext(mockReq.Context())

	err := handler.Handle(mockConn, mockReq)

	if err != nil {
		t.Errorf("Did not expect error but got: %+v", err)
	}

	// assert span attributes
	if span.IsRecording() != false {
		t.Errorf("Since span has ended it should not record anymore")
	}
}

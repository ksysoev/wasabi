package request

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewRetryMiddleware(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	middleware := NewRetryMiddleware(maxRetries, interval)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return fmt.Errorf("mock error")
	})

	ctx := context.Background()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx)

	// Test with successful request
	mockHandlerSuccess := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return nil
	})

	if err := middleware(mockHandlerSuccess).Handle(mockConn, mockReq); err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test with failed request
	if err := middleware(mockHandler).Handle(mockConn, mockReq); err == nil {
		t.Error("Expected error, but got nil")
	}
}

func TestNewRetryMiddleware_CancelledContext(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	middleware := NewRetryMiddleware(maxRetries, interval)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return fmt.Errorf("mock error")
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx)

	// Test with failed request
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	if err != context.Canceled {
		t.Errorf("Expected error to be context.Canceled, but got %v", err)
	}
}

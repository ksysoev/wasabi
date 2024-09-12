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

func TestNewRetryMiddleware_WithLinearRetryPolicy(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	middleware := NewRetryMiddleware(LinearGetRetryInterval(interval), maxRetries)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return fmt.Errorf("mock error")
	})

	ctx := context.Background()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx)

	// Test with successful request
	mockHandlerSuccess := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
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

func TestNewRetryMiddleware_CancelledContext_WithLinearRetryPolicy(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	middleware := NewRetryMiddleware(LinearGetRetryInterval(interval), maxRetries)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
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

func TestNewRetryMiddleware_WithExponentialRetryPolicy(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	delayFactor := 2
	middleware := NewRetryMiddleware(ExponentialGetRetryInterval(interval, delayFactor), maxRetries)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return fmt.Errorf("mock error")
	})

	ctx := context.Background()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx)

	// Test with successful request
	mockHandlerSuccess := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
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

func TestNewRetryMiddleware_CancelledContext_WithExponentialRetryPolicy(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	delayFactor := 2
	middleware := NewRetryMiddleware(ExponentialGetRetryInterval(interval, delayFactor), maxRetries)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
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

func TestNewRetryMiddleware_Predicate(t *testing.T) {
	maxRetries := 3
	interval := time.Microsecond
	delayFactor := 2
	middleware := NewRetryMiddleware(ExponentialGetRetryInterval(interval, delayFactor), maxRetries, func(_ error) bool { return false })

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		return fmt.Errorf("mock error")
	})

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	// Test with failed request
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	if err.Error() != "mock error" {
		t.Errorf("Expected error to be mock error, but got %v", err)
	}
}

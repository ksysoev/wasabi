package request

import (
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewRateLimiterMiddleware(t *testing.T) {
	// Mock requestLimit function
	requestLimit := func(req wasabi.Request) (string, time.Duration, uint64) {
		return "test_key", time.Second, 10
	}

	// Mock next RequestHandler
	next := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		// Mock implementation of next handler
		return nil
	})

	// Create rate limiter middleware
	middleware := NewRateLimiterMiddleware(requestLimit)

	// Create a mock connection and request
	conn := mocks.NewMockConnection(t)
	req := mocks.NewMockRequest(t)

	// Test rate limiter middleware
	err := middleware(next).Handle(conn, req)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

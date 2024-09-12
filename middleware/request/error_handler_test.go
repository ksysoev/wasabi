package request

import (
	"errors"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewErrorHandlingMiddleware(t *testing.T) {
	// Define a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		// Simulate an error
		return errors.New("mock error")
	})

	// Define a mock error handler
	mockErrorHandler := func(conn wasabi.Connection, req wasabi.Request, err error) error {
		// Verify that the error handler is called with the correct parameters
		if conn == nil || req == nil || err == nil {
			t.Error("Error handler called with nil parameters")
		}

		// Return a custom error
		return errors.New("custom error")
	}

	// Create the error handling middleware
	middleware := NewErrorHandlingMiddleware(mockErrorHandler)

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	// Call the middleware with the mock handler
	err := middleware(mockHandler).Handle(mockConn, mockReq)

	// Verify that the error returned by the middleware is the custom error
	if err == nil || err.Error() != "custom error" {
		t.Errorf("Expected error to be 'custom error', but got '%v'", err)
	}
}

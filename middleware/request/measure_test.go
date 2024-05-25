package request

import (
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewMeasurer(t *testing.T) {
	// Define a mock saveMetric function
	var savedMetric wasabi.Request
	var savedError error
	var savedDuration time.Duration
	saveMetric := func(req wasabi.Request, err error, duration time.Duration) {
		savedMetric = req
		savedError = err
		savedDuration = duration
	}

	// Define a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		// Simulate some processing time
		time.Sleep(100 * time.Microsecond)
		return nil
	})

	// Create the measurer middleware
	measurer := NewMeasurer(saveMetric)

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	// Call the measurer middleware with the mock handler
	err := measurer(mockHandler).Handle(mockConn, mockReq)

	// Check if the metric was saved correctly
	if savedMetric != mockReq {
		t.Errorf("Expected saved metric to be %v, but got %v", mockReq, savedMetric)
	}

	if savedError != nil {
		t.Errorf("Expected saved error to be nil, but got %v", savedError)
	}

	if savedDuration < 100*time.Microsecond {
		t.Errorf("Expected saved duration to be greater than 100ms, but got %v", savedDuration)
	}

	// Check if the error is propagated correctly
	if err != nil {
		t.Errorf("Expected error to be nil, but got %v", err)
	}
}

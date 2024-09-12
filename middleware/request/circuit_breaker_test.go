package request

import (
	"fmt"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewCircuitBreakerMiddleware_ClosedState(t *testing.T) {
	threshold := uint32(3)
	period := time.Second

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error { return nil })
	mockRequest := mocks.NewMockRequest(t)
	mockConn := mocks.NewMockConnection(t)

	// Create the circuit breaker middleware
	middleware := NewCircuitBreakerMiddleware(threshold, period)(mockHandler)

	// Test the Closed state
	for i := uint32(0); i < threshold+1; i++ {
		err := middleware.Handle(mockConn, mockRequest)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}
}

func TestNewCircuitBreakerMiddleware_OpenState(t *testing.T) {
	threshold := uint32(1)
	period := time.Second

	testError := fmt.Errorf("test error")

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		time.Sleep(5 * time.Millisecond)
		return testError
	})

	mockRequest := mocks.NewMockRequest(t)
	mockConn := mocks.NewMockConnection(t)

	// Create the circuit breaker middleware
	middleware := NewCircuitBreakerMiddleware(threshold, period)(mockHandler)

	// Bring the circuit breaker to the Open state
	err := middleware.Handle(mockConn, mockRequest)
	if err != testError {
		t.Errorf("Expected error %v, but got %v", testError, err)
	}

	// Test the Open state
	results := make(chan error)

	for i := 0; i < 2; i++ {
		go func() {
			results <- middleware.Handle(mockConn, mockRequest)
		}()
		// Wait out circuit break to change state
		time.Sleep(period)
	}

	OpenErrorCount := 0
	TestErrorCount := 0

	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != ErrCircuitBreakerOpen && err != testError {
				t.Errorf("Expected error %v, but got %v", ErrCircuitBreakerOpen, err)
				continue
			}

			if err == ErrCircuitBreakerOpen {
				OpenErrorCount++
			} else if err == testError {
				TestErrorCount++
			}

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected error, but got none")
		}
	}

	if OpenErrorCount != 1 {
		t.Errorf("Expected 1 ErrCircuitBreakerOpen error, but got %d", OpenErrorCount)
	}

	if TestErrorCount != 1 {
		t.Errorf("Expected 1 test error, but got %d", TestErrorCount)
	}
}

func TestNewCircuitBreakerMiddleware_SemiOpenState(t *testing.T) {
	threshold := uint32(1)
	period := time.Second

	testError := fmt.Errorf("test error")

	errorToReturn := testError

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		time.Sleep(5 * time.Millisecond)
		return errorToReturn
	})

	mockRequest := mocks.NewMockRequest(t)
	mockConn := mocks.NewMockConnection(t)

	// Create the circuit breaker middleware
	middleware := NewCircuitBreakerMiddleware(threshold, period)(mockHandler)

	// Bring the circuit breaker to the Open state
	err := middleware.Handle(mockConn, mockRequest)
	if err != testError {
		t.Errorf("Expected error %v, but got %v", testError, err)
	}

	// Test the Open state
	errorToReturn = nil
	OpenErrorCount := 0
	SuccessCount := 0
	results := make(chan error)

	for i := 0; i < 2; i++ {
		go func() {
			results <- middleware.Handle(mockConn, mockRequest)
		}()
		// Wait out circuit breaker to change state
		time.Sleep(period)
	}

	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != ErrCircuitBreakerOpen && err != nil {
				t.Errorf("Expected error %v, but got %v", ErrCircuitBreakerOpen, err)
				continue
			}

			if err == ErrCircuitBreakerOpen {
				OpenErrorCount++
			} else if err == nil {
				SuccessCount++
			}

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected error, but got none")
		}
	}

	if OpenErrorCount != 1 {
		t.Errorf("Expected 1 ErrCircuitBreakerOpen error, but got %d", OpenErrorCount)
	}

	if SuccessCount != 1 {
		t.Errorf("Expected 1 test error, but got %d", SuccessCount)
	}

	// Confirm that the circuit breaker is now in the Closed state

	for i := 0; i < 2; i++ {
		go func() {
			results <- middleware.Handle(mockConn, mockRequest)
		}()
	}

	OpenErrorCount = 0
	SuccessCount = 0

	for i := 0; i < 2; i++ {
		select {
		case err := <-results:
			if err != ErrCircuitBreakerOpen && err != nil {
				t.Errorf("Expected error %v, but got %v", ErrCircuitBreakerOpen, err)
				continue
			}

			if err == ErrCircuitBreakerOpen {
				OpenErrorCount++
			} else if err == nil {
				SuccessCount++
			}

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected error, but got none")
		}
	}

	if OpenErrorCount != 0 {
		t.Errorf("Expected 0 ErrCircuitBreakerOpen error, but got %d", OpenErrorCount)
	}

	if SuccessCount != 2 {
		t.Errorf("Expected 2 test error, but got %d", SuccessCount)
	}
}

func TestNewCircuitBreakerMiddleware_ResetMeasureInterval(t *testing.T) {
	threshold := uint32(2)
	period := 20 * time.Millisecond

	testError := fmt.Errorf("test error")

	errorToReturn := testError

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		time.Sleep(5 * time.Millisecond)
		return errorToReturn
	})

	mockRequest := mocks.NewMockRequest(t)
	mockConn := mocks.NewMockConnection(t)

	// Create the circuit breaker middleware
	middleware := NewCircuitBreakerMiddleware(threshold, period)(mockHandler)

	// Bring the circuit breaker to the Open state

	for i := uint32(0); i < threshold; i++ {
		if err := middleware.Handle(mockConn, mockRequest); err != testError {
			t.Errorf("Expected error %v, but got %v", testError, err)
		}
	}

	// Confirm that the circuit breaker is now in the Semi-open state
	time.Sleep(period)

	errorToReturn = nil
	results := make(chan error)

	go func() {
		results <- middleware.Handle(mockConn, mockRequest)
	}()

	OpenErrorCount := 0
	SuccessCount := 0

	select {
	case err := <-results:
		if err != ErrCircuitBreakerOpen && err != nil {
			t.Errorf("Expected error %v, but got %v", ErrCircuitBreakerOpen, err)
		}

		if err == ErrCircuitBreakerOpen {
			OpenErrorCount++
		} else if err == nil {
			SuccessCount++
		}

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error, but got none")
	}

	if OpenErrorCount != 0 {
		t.Errorf("Expected 0 ErrCircuitBreakerOpen error, but got %d", OpenErrorCount)
	}

	if SuccessCount != 1 {
		t.Errorf("Expected 1 test error, but got %d", SuccessCount)
	}
}

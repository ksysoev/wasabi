package request

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewTrottlerMiddleware(t *testing.T) {
	limit := uint(3)
	middleware := NewTrottlerMiddleware(limit)

	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		select {
		case <-req.Context().Done():
			return nil
		case <-time.After(time.Second):
			return fmt.Errorf("request timed out")
		}
	})
	ctx1, cancel := context.WithCancel(context.Background())
	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(ctx1)
	// Test with multiple concurrent requests
	wg := sync.WaitGroup{}
	readyChan := make(chan struct{}, limit)

	for i := 0; i < int(limit); i++ {
		wg.Add(1)

		go func() {
			readyChan <- struct{}{}

			err := middleware(mockHandler).Handle(mockConn, mockReq)
			if err != nil && err != context.Canceled {
				t.Errorf("Expected no error, but got %v", err)
			}

			wg.Done()
		}()
	}

	for i := 0; i < int(limit); i++ {
		<-readyChan
	}

	time.Sleep(10 * time.Millisecond)

	mockHandlerInstant := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return nil
	})

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel2()

	mockReq1 := mocks.NewMockRequest(t)
	mockReq1.EXPECT().Context().Return(ctx2)

	// Test with additional requests that should be throttled
	err := middleware(mockHandlerInstant).Handle(mockConn, mockReq1)
	if err == nil {
		t.Error("Expected error due to throttling, but got nil")
	}

	cancel()
	wg.Wait()
}

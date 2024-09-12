package request

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/mock"
)

func TestNewSetTimeoutMiddleware(t *testing.T) {
	timeout := 5 * time.Second
	handler := dispatch.RequestHandlerFunc(func(_ wasabi.Connection, _ wasabi.Request) error {
		// Your custom handler logic here
		return nil
	})

	middleware := NewSetTimeoutMiddleware(timeout)(handler)

	ctx := context.Background()

	conn := mocks.NewMockConnection(t) // Create a mock connection
	req := mocks.NewMockRequest(t)     // Create a mock request

	req.EXPECT().Context().Return(ctx)
	req.EXPECT().WithContext(mock.AnythingOfType("*context.timerCtx")).Return(req)

	err := middleware.Handle(conn, req)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

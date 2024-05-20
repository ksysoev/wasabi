package backend

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"nhooyr.io/websocket"
)

func TestHandle_Success(t *testing.T) {
	expectedData := []byte("request")

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().Send(websocket.MessageText, expectedData).Return(nil)

	r := mocks.NewMockRequest(t)
	r.EXPECT().Context().Return(context.Background())

	// Create a new QueueBackend instance
	onRequest := make(chan string)
	b := NewQueueBackend(func(conn wasabi.Connection, req wasabi.Request, id string) error {
		onRequest <- id
		return nil
	})

	// Call the Handle method
	done := make(chan struct{})
	go func() {
		err := b.Handle(conn, r)
		if err != nil {
			t.Errorf("Unexpected error handling request: %v", err)
		}

		close(done)
	}()

	var reqID string
	select {
	case reqID = <-onRequest:
		if reqID == "" {
			t.Error("Expected request ID to be non-empty")
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected request to be handled")
	}

	b.OnResponse(reqID, websocket.MessageText, expectedData)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected request to be handled")
	}
}

package channel

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"nhooyr.io/websocket"
)

func TestDefaultConnectionRegistry_AddConnection(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	ws, resp, err := websocket.Dial(context.Background(), url, nil)

	if err != nil {
		t.Error(err)
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	ctx := context.Background()

	cb := func(wasabi.Connection, wasabi.MessageType, []byte) {}

	registry := NewDefaultConnectionRegistry()

	conn := registry.AddConnection(ctx, ws, cb)

	if conn == nil {
		t.Error("Expected connection to be created")
	}

	if _, ok := registry.connections[conn.ID()]; !ok {
		t.Error("Expected connection to be added to the registry")
	}
}

func TestDefaultConnectionRegistry_GetConnection(t *testing.T) {
	registry := NewDefaultConnectionRegistry()

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("testID")

	registry.connections[conn.ID()] = conn

	result := registry.GetConnection(conn.ID())

	if result == nil {
		t.Error("Expected connection to be retrieved from the registry")
	}

	if result.ID() != conn.ID() {
		t.Errorf("Expected connection ID to be %s, but got %s", conn.ID(), result.ID())
	}
}

func TestDefaultConnectionRegistry_handleClose(t *testing.T) {
	registry := NewDefaultConnectionRegistry()

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("testID")
	registry.connections[conn.ID()] = conn

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		registry.handleClose()
		wg.Done()
	}()

	registry.onClose <- conn.ID()
	close(registry.onClose)

	wg.Wait()

	if registry.GetConnection(conn.ID()) != nil {
		t.Error("Expected connection to be removed from the registry")
	}
}

func TestDefaultConnectionRegistry_WithMaxFrameLimit(t *testing.T) {
	registry := NewDefaultConnectionRegistry(WithMaxFrameLimit(100))

	if registry.frameSizeLimit != 100 {
		t.Errorf("Unexpected frame size limit: got %d, expected %d", registry.frameSizeLimit, 100)
	}
}

package channel

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"nhooyr.io/websocket"
)

func TestConnectionRegistry_AddConnection(t *testing.T) {
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

	registry := NewConnectionRegistry()

	conn := registry.AddConnection(ctx, ws, cb)

	if conn == nil {
		t.Error("Expected connection to be created")
	}

	if _, ok := registry.connections[conn.ID()]; !ok {
		t.Error("Expected connection to be added to the registry")
	}
}

func TestConnectionRegistry_AddConnection_ToClosedRegistry(t *testing.T) {
	registry := NewConnectionRegistry()

	registry.Close()

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

	conn := registry.AddConnection(ctx, ws, cb)

	if conn != nil {
		t.Error("Expected connection to be nil")
	}
}

func TestConnectionRegistry_GetConnection(t *testing.T) {
	registry := NewConnectionRegistry()

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

func TestConnectionRegistry_handleClose(t *testing.T) {
	registry := NewConnectionRegistry()

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

func TestConnectionRegistry_WithMaxFrameLimit(t *testing.T) {
	registry := NewConnectionRegistry(WithMaxFrameLimit(100))

	if registry.frameSizeLimit != 100 {
		t.Errorf("Unexpected frame size limit: got %d, expected %d", registry.frameSizeLimit, 100)
	}
}
func TestConnectionRegistry_Shutdown(t *testing.T) {
	ctx := context.Background()
	registry := NewConnectionRegistry()

	// Add some mock connections to the registry
	conn1 := mocks.NewMockConnection(t)
	conn2 := mocks.NewMockConnection(t)

	conn1.EXPECT().ID().Return("conn1")
	conn2.EXPECT().ID().Return("conn2")

	registry.connections[conn1.ID()] = conn1
	registry.connections[conn2.ID()] = conn2

	// Set up expectations for the Close method
	conn1.EXPECT().Close(websocket.StatusServiceRestart, "", ctx).Return(nil)
	conn2.EXPECT().Close(websocket.StatusServiceRestart, "", ctx).Return(nil)

	err := registry.Close(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that the registry is closed
	if !registry.isClosed {
		t.Error("Expected registry to be closed")
	}
}
func TestConnectionRegistry_WithConcurrencyLimit(t *testing.T) {
	registry := NewConnectionRegistry()

	if registry.concurrencyLimit != concurencyLimitPerConnection {
		t.Errorf("Unexpected concurrency limit: got %d, expected %d", registry.concurrencyLimit, concurencyLimitPerConnection)
	}

	registry = NewConnectionRegistry(WithConcurrencyLimit(10))

	if registry.concurrencyLimit != 10 {
		t.Errorf("Unexpected concurrency limit: got %d, expected %d", registry.concurrencyLimit, 10)
	}
}
func TestConnectionRegistry_WithInActivityTimeout(t *testing.T) {
	registry := NewConnectionRegistry()

	if registry.inActivityTimeout != 0 {
		t.Errorf("Unexpected inactivity timeout: got %v, expected %v", registry.inActivityTimeout, 0)
	}

	registry = NewConnectionRegistry(WithInActivityTimeout(5 * time.Minute))

	if registry.inActivityTimeout != 5*time.Minute {
		t.Errorf("Unexpected inactivity timeout: got %s, expected %s", registry.inActivityTimeout, 5*time.Minute)
	}
}
func TestConnectionRegistry_WithOnConnect(t *testing.T) {
	registry := NewConnectionRegistry()

	if registry.onConnect != nil {
		t.Error("Expected onConnect callback to be nil")
	}

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

	executed := false
	cb := func(conn wasabi.Connection) {
		if conn == nil {
			t.Error("Expected connection to be passed to onConnect callback")
		}

		executed = true
	}

	registry = NewConnectionRegistry(WithOnConnectHook(cb))

	if registry.onConnect == nil {
		t.Error("Expected onConnect callback to be set")
	}

	registry.AddConnection(context.Background(), ws, func(wasabi.Connection, wasabi.MessageType, []byte) {})

	if !executed {
		t.Error("Expected onConnect callback to be executed")
	}
}

func TestConnectionRegistry_WithOnDisconnectHook(t *testing.T) {
	registry := NewConnectionRegistry()

	if registry.onDisconnect != nil {
		t.Error("Expected onDisconnect hook to be nil")
	}

	done := make(chan struct{})
	hook := func(conn wasabi.Connection) {
		if conn == nil {
			t.Error("Expected connection to be passed to onDisconnect hook")
		}

		done <- struct{}{}
	}

	registry = NewConnectionRegistry(WithOnDisconnectHook(hook))

	if registry.onDisconnect == nil {
		t.Error("Expected onDisconnect hook to be set")
	}

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
	conn := registry.AddConnection(ctx, ws, cb)

	registry.onClose <- conn.ID()
	close(registry.onClose)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected onDisconnect hook to be executed")
	}
}

package channel

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestConnectionRegistry_HandleConnection(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	ws, resp, err := websocket.Dial(context.Background(), url, nil)

	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %v", err)
	}

	err = ws.Write(context.Background(), websocket.MessageText, []byte("test"))
	if err != nil {
		t.Errorf("Unexpected error writing to websocket: %v", err)
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	ready := make(chan struct{})
	cb := func(wasabi.Connection, wasabi.MessageType, []byte) {
		close(ready)
	}

	registry := NewConnectionRegistry()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		registry.HandleConnection(ctx, ws, cb)
		close(done)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be handled")
	}

	if len(registry.connections) != 1 {
		t.Error("Expected connection to be added to the registry")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be closed")
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

	done := make(chan struct{})
	go func() {
		registry.HandleConnection(ctx, ws, cb)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be closed")
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		registry.HandleConnection(ctx, ws, func(wasabi.Connection, wasabi.MessageType, []byte) {})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be closed")
	}

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

		close(done)
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

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cb := func(wasabi.Connection, wasabi.MessageType, []byte) {}

	ready := make(chan struct{})

	go func() {
		registry.HandleConnection(ctx, ws, cb)
		close(ready)
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be handled")
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected onDisconnect hook to be executed")
	}
}

func TestConnectionRegistry_WithConnectionLimit(t *testing.T) {
	registry := NewConnectionRegistry()

	if registry.connectionLimit != -1 {
		t.Errorf("Unexpected connection limit: got %d, expected %d", registry.connectionLimit, -1)
	}

	registry = NewConnectionRegistry(WithConnectionLimit(10))

	if registry.connectionLimit != 10 {
		t.Errorf("Unexpected connection limit: got %d, expected %d", registry.connectionLimit, 10)
	}
}

func TestConnectionRegistry_AddConnection_ConnectionLimitReached(t *testing.T) {
	registry := NewConnectionRegistry(WithConnectionLimit(2))
	conn1 := mocks.NewMockConnection(t)
	conn2 := mocks.NewMockConnection(t)

	conn1.EXPECT().ID().Return("conn1")
	conn2.EXPECT().ID().Return("conn2")

	registry.connections[conn1.ID()] = conn1
	registry.connections[conn2.ID()] = conn2

	ctx := context.Background()
	cb := func(wasabi.Connection, wasabi.MessageType, []byte) {}

	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	ws, resp, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %v", err)
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	done := make(chan struct{})
	go func() {
		registry.HandleConnection(ctx, ws, cb)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected connection to be handled")
	}

	if len(registry.connections) != 2 {
		t.Error("Expected connection to not be added to the registry")
	}
}

func TestConnectionRegistry_CanAccept_ConnectionLimitNotSet(t *testing.T) {
	registry := NewConnectionRegistry()

	if !registry.CanAccept() {
		t.Error("Expected CanAccept to return true when connection limit is not set")
	}

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("conn1")

	registry.connections[conn.ID()] = conn

	if !registry.CanAccept() {
		t.Error("Expected CanAccept to return true when connection limit is not set")
	}
}

func TestConnectionRegistry_CanAccept_ConnectionLimitReached(t *testing.T) {
	registry := NewConnectionRegistry(WithConnectionLimit(2))

	conn1 := mocks.NewMockConnection(t)
	conn1.EXPECT().ID().Return("conn1")

	registry.connections[conn1.ID()] = conn1

	if !registry.CanAccept() {
		t.Error("Expected CanAccept to return true when connection limit is reached")
	}

	conn2 := mocks.NewMockConnection(t)
	conn2.EXPECT().ID().Return("conn2")
	registry.connections[conn2.ID()] = conn2

	if registry.CanAccept() {
		t.Error("Expected CanAccept to return false when connection limit is reached")
	}
}

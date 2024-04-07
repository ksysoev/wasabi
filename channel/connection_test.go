package channel

import (
	"context"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"golang.org/x/net/websocket"
)

func TestConn_ID(t *testing.T) {
	ws := &websocket.Conn{}
	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose)

	if conn.ID() == "" {
		t.Error("Expected connection ID to be non-empty")
	}
}

func TestConn_Context(t *testing.T) {
	ws := &websocket.Conn{}
	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose)

	if conn.Context() == nil {
		t.Error("Expected connection context to be non-nil")
	}
}

func TestConn_HandleRequests(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) { _, _ = io.Copy(ws, ws) }))
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()

	ws, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %v", err)
	}

	defer ws.Close()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose)

	// Mock OnMessage callback
	var wg sync.WaitGroup

	wg.Add(1)

	conn.onMessageCB = func(c wasabi.Connection, data []byte) { wg.Done() }

	go conn.HandleRequests()

	// Send message to trigger OnMessage callback
	err = websocket.Message.Send(ws, []byte("test message"))
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}

	wg.Wait()
}

func TestConn_Send(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) { _, _ = io.Copy(ws, ws) }))
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	ws, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %v", err)
	}

	defer ws.Close()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose)

	err = conn.Send([]byte("test message"))
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}
}

func TestConn_close(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) { _, _ = io.Copy(ws, ws) }))
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	ws, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %v", err)
	}

	defer ws.Close()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose)

	// Mock OnClose channel
	closeChan := make(chan string)
	conn.onClose = closeChan

	go conn.close()

	select {
	case id, ok := <-closeChan:
		if !ok {
			t.Error("Expected OnClose channel to be closed")
		}

		if id != conn.ID() {
			t.Errorf("Expected ID to be %s, but got %s", conn.ID(), id)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected OnClose channel to be called")
	}
}

package channel

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

var wsHandlerEcho = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	for {
		_, wsr, err := c.Reader(r.Context())
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}

		wsw, err := c.Writer(r.Context(), websocket.MessageText)
		if err != nil {
			return
		}

		_, err = io.Copy(wsw, wsr)
		if err != nil {
			return
		}

		err = wsw.Close()
		if err != nil {
			return
		}
	}
})

func TestConn_ID(t *testing.T) {
	ws := &websocket.Conn{}
	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose, newBufferPool(), 1)

	if conn.ID() == "" {
		t.Error("Expected connection ID to be non-empty")
	}
}

func TestConn_Context(t *testing.T) {
	ws := &websocket.Conn{}
	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose, newBufferPool(), 1)

	if conn.Context() == nil {
		t.Error("Expected connection context to be non-nil")
	}
}

func TestConn_HandleRequests(t *testing.T) {
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

	defer func() { _ = ws.CloseNow() }()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose, newBufferPool(), 1)

	// Mock OnMessage callback
	received := make(chan struct{})

	conn.onMessageCB = func(c wasabi.Connection, msgType wasabi.MessageType, data []byte) { received <- struct{}{} }

	go conn.HandleRequests()

	// Send message to trigger OnMessage callback
	err = ws.Write(context.Background(), websocket.MessageText, []byte("test message"))
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}

	select {
	case <-received:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Error("Expected OnMessage callback to be called")
	}
}

func TestConn_Send(t *testing.T) {
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

	defer func() { _ = ws.CloseNow() }()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose, newBufferPool(), 1)

	err = conn.Send(wasabi.MsgTypeText, []byte("test message"))
	if err != nil {
		t.Errorf("Unexpected error sending message: %v", err)
	}
}

func TestConn_close(t *testing.T) {
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

	defer func() { _ = ws.CloseNow() }()

	onClose := make(chan string)
	conn := NewConnection(context.Background(), ws, nil, onClose, newBufferPool(), 1)

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

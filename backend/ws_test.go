package backend

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksysoev/wasabi/mocks"
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

func TestNewWSBackend(t *testing.T) {
	url := "ws://example.com"
	b := NewWSBackend(url)

	if b.URL != url {
		t.Errorf("Expected URL to be %q, but got %q", url, b.URL)
	}

	if b.connections == nil {
		t.Error("Expected connections map to be initialized, but got nil")
	}

	if b.lock == nil {
		t.Error("Expected lock to be initialized, but got nil")
	}
}

func TestGetConnectionExistingConnection(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	b := NewWSBackend(url)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("connection1")

	c := &websocket.Conn{}
	b.connections[conn.ID()] = c

	got, err := b.getConnection(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got != c {
		t.Errorf("Expected existing connection, but got different connection")
	}
}

func TestGetConnectionNewConnection(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	defer server.Close()
	url := "ws://" + server.Listener.Addr().String()

	b := NewWSBackend(url)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("connection1")
	conn.EXPECT().Context().Return(context.Background())

	got, err := b.getConnection(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got == nil {
		t.Errorf("Expected new connection, but got nil")
	}

	if b.connections[conn.ID()] != got {
		t.Errorf("Expected connection to be stored in the connections map")
	}
}

func TestGetConnectionDialError(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	url := "ws://" + server.Listener.Addr().String()
	server.Close()

	b := NewWSBackend(url)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("connection1")
	conn.EXPECT().Context().Return(context.Background())

	_, err := b.getConnection(conn)
	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}

	if b.connections[conn.ID()] != nil {
		t.Errorf("Expected connection to not be stored in the connections map")
	}
}

func TestWSBackend_Handle(t *testing.T) {
	server := httptest.NewServer(wsHandlerEcho)
	url := "ws://" + server.Listener.Addr().String()

	defer server.Close()

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().ID().Return("connection1")
	conn.EXPECT().Context().Return(context.Background())
	waitForSend := conn.
		EXPECT().
		Send(websocket.MessageText, []byte("Hello, world!")).
		Return(nil).
		Once().
		WaitUntil(time.After(50 * time.Millisecond)).WaitFor

	r := mocks.NewMockRequest(t)
	r.EXPECT().Data().Return([]byte("Hello, world!"))
	r.EXPECT().Context().Return(context.Background())

	b := NewWSBackend(url)

	err := b.Handle(conn, r)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	<-waitForSend
}

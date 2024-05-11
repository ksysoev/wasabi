package backend

import (
	"bytes"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type WSBackend struct {
	URL         string
	connections map[string]*websocket.Conn
	lock        *sync.RWMutex
}

// NewWSBackend creates a new instance of WSBackend with the specified URL.
func NewWSBackend(url string) *WSBackend {
	return &WSBackend{
		URL:         url,
		connections: make(map[string]*websocket.Conn),
		lock:        &sync.RWMutex{},
	}
}

// Handle handles the incoming request from the WebSocket connection.
// It writes the request data to the WebSocket connection's context.
// The function returns an error if there is any issue with the connection or writing the data.
func (b *WSBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	c, err := b.getConnection(conn)

	if err != nil {
		return err
	}

	return c.Write(r.Context(), websocket.MessageText, r.Data())
}

// getConnection returns the websocket connection associated with the given connection.
// If the connection is already established, it returns the existing connection.
// Otherwise, it establishes a new connection and returns it.
func (b *WSBackend) getConnection(conn wasabi.Connection) (*websocket.Conn, error) {
	b.lock.RLock()
	c, ok := b.connections[conn.ID()]
	b.lock.RUnlock()

	if ok {
		return c, nil
	}

	c, _, err := websocket.Dial(conn.Context(), b.URL, nil)
	if err != nil {
		return nil, err
	}

	b.lock.Lock()
	b.connections[conn.ID()] = c

	go b.responseHandler(c, conn)
	b.lock.Unlock()

	return c, nil
}

// responseHandler handles the response from the server to the client.
// It reads messages from the server, sends them to the client, and manages the connection lifecycle.
func (b *WSBackend) responseHandler(server *websocket.Conn, client wasabi.Connection) {
	defer func() {
		b.lock.Lock()
		delete(b.connections, client.ID())
		server.Close(websocket.StatusNormalClosure, "")
		client.Close(websocket.StatusNormalClosure, "")
		b.lock.Unlock()
	}()

	buffer := bytes.NewBuffer(make([]byte, 0))
	ctx := client.Context()

	for ctx.Err() == nil {
		buffer.Reset()
		msgType, reader, err := server.Reader(ctx)
		if err != nil {
			return
		}

		_, err = buffer.ReadFrom(reader)

		if err != nil {
			return
		}

		err = client.Send(msgType, buffer.Bytes())
		if err != nil {
			return
		}
	}
}

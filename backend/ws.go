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

func NewWSBackend(url string) *WSBackend {
	return &WSBackend{
		URL:         url,
		connections: make(map[string]*websocket.Conn),
		lock:        &sync.RWMutex{},
	}
}

func (b *WSBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	c, err := b.getConnection(conn)

	if err != nil {
		return err
	}

	return c.Write(r.Context(), websocket.MessageText, r.Data())
}

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

	go func() {
		defer func() {
			b.lock.Lock()
			delete(b.connections, conn.ID())
			conn.Close(websocket.StatusNormalClosure, "")
			b.lock.Unlock()
		}()
		buffer := bytes.NewBuffer(make([]byte, 0))
		for {
			buffer.Reset()
			msgType, reader, err := c.Reader(conn.Context())
			if err != nil {
				return
			}

			_, err = buffer.ReadFrom(reader)

			if err != nil {
				return
			}

			err = conn.Send(msgType, buffer.Bytes())
			if err != nil {
				return
			}
		}
	}()
	b.lock.Unlock()

	return c, nil
}

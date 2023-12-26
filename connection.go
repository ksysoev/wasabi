package wasabi

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

type Connection struct {
	ws          *websocket.Conn
	repsChan    chan string
	isClosed    atomic.Bool
	waitGroup   *sync.WaitGroup
	onMessageCB onMessage
}

type onMessage func(conn *Connection, msg string) error

func NewConnection(ws *websocket.Conn, cb onMessage) *Connection {
	conn := &Connection{
		ws:          ws,
		onMessageCB: cb,
		waitGroup:   &sync.WaitGroup{},
	}

	return conn
}

func (c *Connection) Close() {
	if c.isClosed.Load() {
		return
	}

	c.isClosed.Store(true)

	c.ws.Close()
}

func (c *Connection) HandleRequest() {
	for {
		var msg string
		err := websocket.Message.Receive(c.ws, &msg)

		if err != nil {
			if c.isClosed.Load() {
				return
			}

			if err.Error() == "EOF" {
				slog.Debug("Connection closed")
				c.Close()
				return
			}

			slog.Debug("Error reading message: " + err.Error())
			continue
		}

		c.onMessage(msg)
	}
}

func (c *Connection) onMessage(msg string) {
	slog.Debug("Received message: " + msg)
	go c.onMessageCB(c, msg)
}

func (c *Connection) SendResponse(msg string) error {
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	if c.isClosed.Load() {
		return nil
	}

	return websocket.Message.Send(c.ws, msg)
}

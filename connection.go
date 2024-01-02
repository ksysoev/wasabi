package wasabi

import (
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type ConnectionRegistry interface {
	AddConnection(conn *Connection) *Connection
	GetConnection(id string) *Connection
}

type DefaultConnectionRegistry struct {
	connections map[string]*Connection
	mu          sync.RWMutex
	onClose     chan string
}

func NewDefaultConnectionRegistry() *DefaultConnectionRegistry {
	reg := &DefaultConnectionRegistry{
		connections: make(map[string]*Connection),
		onClose:     make(chan string),
	}

	go reg.handleClose()

	return reg
}

func (r *DefaultConnectionRegistry) AddConnection(conn *Connection) *Connection {
	conn.onClose = r.onClose
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connections[conn.ID()] = conn

	return conn
}

func (r *DefaultConnectionRegistry) GetConnection(id string) *Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connections[id]
}

func (r *DefaultConnectionRegistry) handleClose() {
	for id := range r.onClose {
		r.mu.Lock()
		delete(r.connections, id)
		r.mu.Unlock()
	}
}

type Connection struct {
	id          string
	ws          *websocket.Conn
	repsChan    chan string
	isClosed    atomic.Bool
	waitGroup   *sync.WaitGroup
	onMessageCB onMessage
	onClose     chan<- string
	stash       Stasher
}

type onMessage func(conn *Connection, req Request) error

func NewConnection() *Connection {
	conn := &Connection{
		id:        uuid.New().String(),
		waitGroup: &sync.WaitGroup{},
		stash:     NewStashStore(),
	}

	return conn
}

func (c *Connection) Stash() Stasher {
	return c.stash
}

func (c *Connection) ID() string {
	return c.id
}

func (c *Connection) Close() {
	if c.isClosed.Load() {
		return
	}

	c.onClose <- c.id
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

	req, err := NewRequest(msg)
	if err != nil {
		slog.Debug("Error parsing request: " + err.Error())
		resp := ResponseFromError(err)

		data, err := resp.String()
		if err != nil {
			slog.Debug("Error creating response: " + err.Error())
			return
		}

		c.SendResponse(data)
		return
	}

	go c.onMessageCB(c, req)
}

func (c *Connection) SendResponse(msg string) error {
	c.waitGroup.Add(1)
	defer c.waitGroup.Done()

	if c.isClosed.Load() {
		return nil
	}

	return websocket.Message.Send(c.ws, msg)
}

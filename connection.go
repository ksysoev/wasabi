package wasabi

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

// ConnectionRegistry is interface for connection registries
type ConnectionRegistry interface {
	AddConnection(
		ctx context.Context,
		ws *websocket.Conn,
		cb onMessage,
	) Connection
	GetConnection(id string) Connection
}

// DefaultConnectionRegistry is default implementation of ConnectionRegistry
type DefaultConnectionRegistry struct {
	connections map[string]Connection
	onClose     chan string
	mu          sync.RWMutex
}

// NewDefaultConnectionRegistry creates new instance of DefaultConnectionRegistry
func NewDefaultConnectionRegistry() *DefaultConnectionRegistry {
	reg := &DefaultConnectionRegistry{
		connections: make(map[string]Connection),
		onClose:     make(chan string),
	}

	go reg.handleClose()

	return reg
}

// AddConnection adds new Websocket connection to registry
func (r *DefaultConnectionRegistry) AddConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb onMessage,
) Connection {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn := NewConnection(ctx, ws, cb, r.onClose)
	r.connections[conn.ID()] = conn

	return conn
}

// GetConnection returns connection by id
func (r *DefaultConnectionRegistry) GetConnection(id string) Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connections[id]
}

// handleClose handles connection cloasures and removes them from registry
func (r *DefaultConnectionRegistry) handleClose() {
	for id := range r.onClose {
		r.mu.Lock()
		delete(r.connections, id)
		r.mu.Unlock()
	}
}

// Connection is interface for connections
type Connection interface {
	Send(msg []byte) error
	Context() context.Context
	ID() string
	HandleRequests()
}

// Conn is default implementation of Connection
type Conn struct {
	ws          *websocket.Conn
	ctx         context.Context
	onMessageCB onMessage
	onClose     chan<- string
	ctxCancel   context.CancelFunc
	id          string
	isClosed    atomic.Bool
}

// onMessage is type for onMessage callback
type onMessage func(conn Connection, data []byte)

// NewConnection creates new instance of websocket connection
func NewConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb onMessage,
	onClose chan<- string,
) *Conn {
	ctx, cancel := context.WithCancel(ctx)

	return &Conn{
		ws:          ws,
		id:          uuid.New().String(),
		ctx:         ctx,
		ctxCancel:   cancel,
		onMessageCB: cb,
		onClose:     onClose,
	}
}

// ID returns connection id
func (c *Conn) ID() string {
	return c.id
}

// Context returns connection context
func (c *Conn) Context() context.Context {
	return c.ctx
}

// HandleRequests handles incoming messages
func (c *Conn) HandleRequests() {
	defer c.close()

	for c.ctx.Err() == nil {
		var data []byte
		err := websocket.Message.Receive(c.ws, &data)

		if err != nil {
			if c.isClosed.Load() {
				return
			}

			if err.Error() == "EOF" {
				slog.Debug("Connection closed")
				c.close()

				return
			}

			if err.Error() == "ErrFrameTooLarge" {
				// Unexpectedkly large message received
				// it's probably more safe to close connection
				c.close()

				return
			}

			slog.Info("Error reading message: " + err.Error())

			continue
		}

		go c.onMessageCB(c, data)
	}
}

// Send sends message to connection
func (c *Conn) Send(msg []byte) error {
	if c.isClosed.Load() || c.ctx.Err() != nil {
		return fmt.Errorf("connection is closed")
	}

	return websocket.Message.Send(c.ws, msg)
}

// close closes connection
func (c *Conn) close() {
	if c.isClosed.Load() {
		return
	}

	c.ctxCancel()
	c.onClose <- c.id
	c.isClosed.Store(true)

	c.ws.Close()
}

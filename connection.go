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

type ConnectionRegistry interface {
	AddConnection(
		ctx context.Context,
		ws *websocket.Conn,
		cb onMessage,
	) Connection
	GetConnection(id string) Connection
}

type DefaultConnectionRegistry struct {
	connections map[string]Connection
	mu          sync.RWMutex
	onClose     chan string
}

func NewDefaultConnectionRegistry() *DefaultConnectionRegistry {
	reg := &DefaultConnectionRegistry{
		connections: make(map[string]Connection),
		onClose:     make(chan string),
	}

	go reg.handleClose()

	return reg
}

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

func (r *DefaultConnectionRegistry) GetConnection(id string) Connection {
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

type Connection interface {
	Send(msg []byte) error
	Context() context.Context
	ID() string
	SetOnClose(chan<- string)
	HandleRequests()
}

type Conn struct {
	id          string
	ws          *websocket.Conn
	ctx         context.Context
	isClosed    atomic.Bool
	onMessageCB onMessage
	onClose     chan<- string
	ctxCancel   context.CancelFunc
}

type onMessage func(conn Connection, data []byte)

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

func (c *Conn) ID() string {
	return c.id
}

func (c *Conn) Close() {
	if c.isClosed.Load() {
		return
	}

	c.ctxCancel()
	c.onClose <- c.id
	c.isClosed.Store(true)

	c.ws.Close()
}

func (c *Conn) Context() context.Context {
	return c.ctx
}

func (c *Conn) SetOnClose(onClose chan<- string) {
	c.onClose = onClose
}

func (c *Conn) HandleRequests() {
	defer c.Close()

	for c.ctx.Err() == nil {
		var data []byte
		err := websocket.Message.Receive(c.ws, &data)

		if err != nil {
			if c.isClosed.Load() {
				return
			}

			if err.Error() == "EOF" {
				slog.Debug("Connection closed")
				c.Close()
				return
			}

			if err.Error() == "ErrFrameTooLarge" {
				// Unexpectedkly large message received
				// it's probably more safe to close connection
				c.Close()
				return
			}

			slog.Info("Error reading message: " + err.Error())
			continue
		}
		go c.onMessageCB(c, data)
	}
}

func (c *Conn) Send(msg []byte) error {
	if c.isClosed.Load() || c.ctx.Err() != nil {
		return fmt.Errorf("connection is closed")
	}

	return websocket.Message.Send(c.ws, msg)
}

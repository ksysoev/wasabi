package channel

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"

	"github.com/ksysoev/wasabi"
)

var (
	// ErrConnectionClosed is error for closed connections
	ErrConnectionClosed = errors.New("connection is closed")
)

// Conn is default implementation of Connection
type Conn struct {
	ws          *websocket.Conn
	reqWG       *sync.WaitGroup
	ctx         context.Context
	onMessageCB wasabi.OnMessage
	onClose     chan<- string
	ctxCancel   context.CancelFunc
	id          string
	isClosed    atomic.Bool
}

// NewConnection creates new instance of websocket connection
func NewConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
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
		reqWG:       &sync.WaitGroup{},
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
			switch {
			case c.isClosed.Load():
				return
			case errors.Is(err, io.EOF):
				return
			case errors.Is(err, websocket.ErrFrameTooLarge):
				return
			case errors.Is(err, net.ErrClosed):
				return
			default:
				slog.Warn("Error reading message: ", err)
			}

			slog.Warn("Error reading message: " + err.Error())

			continue
		}

		c.reqWG.Add(1)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			c.onMessageCB(c, data)
		}(c.reqWG)
	}
}

// Send sends message to connection
func (c *Conn) Send(msg any) error {
	if c.isClosed.Load() || c.ctx.Err() != nil {
		return ErrConnectionClosed
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
	c.reqWG.Wait()
}

package channel

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"github.com/ksysoev/wasabi"
)

var (
	// ErrConnectionClosed is error for closed connections
	ErrConnectionClosed = errors.New("connection is closed")
)

// Conn is default implementation of Connection
type Conn struct {
	ctx         context.Context
	ws          *websocket.Conn
	reqWG       *sync.WaitGroup
	onMessageCB wasabi.OnMessage
	onClose     chan<- string
	ctxCancel   context.CancelFunc
	bufferPool  *bufferPool
	sem         chan struct{}
	id          string
	isClosed    atomic.Bool
}

// NewConnection creates new instance of websocket connection
func NewConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
	onClose chan<- string,
	bufferPool *bufferPool,
	concurrencyLimit uint,
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
		bufferPool:  bufferPool,
		sem:         make(chan struct{}, concurrencyLimit),
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
		c.sem <- struct{}{}

		buffer := c.bufferPool.get()
		msgType, reader, err := c.ws.Reader(c.ctx)

		if err != nil {
			return
		}

		_, err = buffer.ReadFrom(reader)
		if err != nil {
			switch {
			case errors.Is(err, io.EOF):
				return
			case errors.Is(err, context.Canceled):
				return
			}

			slog.Warn("Error reading message: " + err.Error())

			return
		}

		c.reqWG.Add(1)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			c.onMessageCB(c, msgType, buffer.Bytes())
			c.bufferPool.put(buffer)
			<-c.sem
		}(c.reqWG)
	}
}

// Send sends message to connection
func (c *Conn) Send(msgType wasabi.MessageType, msg []byte) error {
	if c.isClosed.Load() || c.ctx.Err() != nil {
		return ErrConnectionClosed
	}

	return c.ws.Write(c.ctx, msgType, msg)
}

// close closes the connection.
// It cancels the context, sends the connection ID to the onClose channel,
// marks the connection as closed, and waits for any pending requests to complete.
func (c *Conn) close() {
	if c.isClosed.Load() {
		return
	}

	c.ctxCancel()
	c.onClose <- c.id
	c.isClosed.Store(true)

	_ = c.ws.CloseNow()
	c.reqWG.Wait()
}

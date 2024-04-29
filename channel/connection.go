package channel

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"github.com/ksysoev/wasabi"
)

var (
	// ErrConnectionClosed is error for closed connections
	ErrConnectionClosed = errors.New("connection is closed")
)

type state int32

const (
	connected  state = iota // initial and normal state of the connection
	closing                 // connection is closing, it means that we stop accepting new requests but let the existing ones to finish
	terminated              // connection is closed
)

// Conn is default implementation of Connection
type Conn struct {
	ctx             context.Context
	ws              *websocket.Conn
	reqWG           *sync.WaitGroup
	onMessageCB     wasabi.OnMessage
	onClose         chan<- string
	ctxCancel       context.CancelFunc
	bufferPool      *bufferPool
	state           *atomic.Int32
	sem             chan struct{}
	inActiveTimer   *time.Timer
	id              string
	inActiveTimeout time.Duration
}

// NewConnection creates new instance of websocket connection
func NewConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
	onClose chan<- string,
	bufferPool *bufferPool,
	concurrencyLimit uint,
	inActivityTimeout time.Duration,
) *Conn {
	ctx, cancel := context.WithCancel(ctx)
	state := atomic.Int32{}
	state.Store(int32(connected))

	conn := &Conn{
		ws:              ws,
		id:              uuid.New().String(),
		ctx:             ctx,
		ctxCancel:       cancel,
		onMessageCB:     cb,
		onClose:         onClose,
		reqWG:           &sync.WaitGroup{},
		state:           &state,
		bufferPool:      bufferPool,
		sem:             make(chan struct{}, concurrencyLimit),
		inActiveTimeout: inActivityTimeout,
	}

	if conn.inActiveTimeout > 0 {
		conn.inActiveTimer = time.NewTimer(conn.inActiveTimeout)
		go conn.watchInactivity()
	}

	return conn
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

		if c.inActiveTimeout > 0 {
			c.inActiveTimer.Reset(c.inActiveTimeout)
		}

		buffer := c.bufferPool.get()
		msgType, reader, err := c.ws.Reader(c.ctx)

		if err != nil {
			return
		}

		_, err = buffer.ReadFrom(reader)

		if c.state.Load() == int32(closing) {
			continue
		}

		if err != nil {
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, net.ErrClosed):
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
	if c.ctx.Err() != nil {
		return ErrConnectionClosed
	}

	if c.inActiveTimeout > 0 {
		c.inActiveTimer.Reset(c.inActiveTimeout)
	}

	return c.ws.Write(c.ctx, msgType, msg)
}

// close closes the connection.
// It cancels the context, sends the connection ID to the onClose channel,
// marks the connection as closed, and waits for any pending requests to complete.
func (c *Conn) close() {
	if !c.state.CompareAndSwap(int32(connected), int32(terminated)) &&
		!c.state.CompareAndSwap(int32(closing), int32(terminated)) {
		return
	}

	c.ctxCancel()
	c.onClose <- c.id

	// Terminate the connection immediately.
	_ = c.ws.CloseNow()

	// Wait for any pending requests to complete.
	c.reqWG.Wait()
}

// Close closes the connection with the specified status code and reason.
// If the connection is already closed, it returns an error.
// If a closing context is provided, it waits for pending requests to complete
// before closing the connection. If the context is canceled, the connection
// is closed immediately. If there are no pending requests, the connection is
// closed immediately. After closing the connection, the connection state is
// set to terminated and the `onClose` channel is notified with the connection ID.
func (c *Conn) Close(status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	if !c.state.CompareAndSwap(int32(connected), int32(closing)) {
		return ErrConnectionClosed
	}

	if len(closingCtx) > 0 {
		ctx := closingCtx[0]
		done := make(chan struct{})

		go func() {
			c.reqWG.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done(): // If the context is canceled, we should close the connection immediately.
		case <-done: // If there are no pending requests, we can close the connection immediately.
		case <-c.ctx.Done(): // If the connection is already closed, we should not wait for pending requests.
		}
	}

	_ = c.ws.Close(status, reason)

	c.ctxCancel()
	c.state.Store(int32(terminated))
	c.onClose <- c.id

	return nil
}

// watchInactivity watches for inactivity on the connection.
// If the connection is inactive for the specified duration, it closes the connection.
func (c *Conn) watchInactivity() {
	defer c.inActiveTimer.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.inActiveTimer.C:
			_ = c.Close(websocket.StatusGoingAway, "inactivity timeout")
			return
		}
	}
}

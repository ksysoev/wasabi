package channel

import (
	"context"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

const (
	concurencyLimitPerConnection = 25
	frameSizeLimitInBytes        = 32768
)

// ConnectionRegistry is default implementation of ConnectionRegistry
type ConnectionRegistry struct {
	connections      map[string]wasabi.Connection
	onClose          chan string
	bufferPool       *bufferPool
	concurrencyLimit uint
	mu               sync.RWMutex
	frameSizeLimit   int64
	isClosed         bool
}

type ConnectionRegistryOption func(*ConnectionRegistry)

// NewConnectionRegistry creates new instance of ConnectionRegistry
func NewConnectionRegistry(opts ...ConnectionRegistryOption) *ConnectionRegistry {
	reg := &ConnectionRegistry{
		connections:      make(map[string]wasabi.Connection),
		onClose:          make(chan string),
		concurrencyLimit: concurencyLimitPerConnection,
		bufferPool:       newBufferPool(),
		frameSizeLimit:   frameSizeLimitInBytes,
		isClosed:         false,
	}

	for _, opt := range opts {
		opt(reg)
	}

	go reg.handleClose()

	return reg
}

// AddConnection adds new Websocket connection to registry
func (r *ConnectionRegistry) AddConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
) wasabi.Connection {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isClosed {
		return nil
	}

	conn := NewConnection(ctx, ws, cb, r.onClose, r.bufferPool, r.concurrencyLimit)
	r.connections[conn.ID()] = conn

	conn.ws.SetReadLimit(r.frameSizeLimit)

	return conn
}

// GetConnection returns connection by id
func (r *ConnectionRegistry) GetConnection(id string) wasabi.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connections[id]
}

// handleClose handles connection cloasures and removes them from registry
func (r *ConnectionRegistry) handleClose() {
	for id := range r.onClose {
		r.mu.Lock()
		delete(r.connections, id)
		r.mu.Unlock()
	}
}

// Shutdown closes all connections in the ConnectionRegistry.
// It sets the isClosed flag to true, indicating that the registry is closed.
// It then iterates over all connections, closes them with the given context,
// and waits for all closures to complete before returning.
func (r *ConnectionRegistry) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	r.isClosed = true
	connections := make([]wasabi.Connection, 0, len(r.connections))

	for _, conn := range r.connections {
		connections = append(connections, conn)
	}

	r.mu.Unlock()

	wg := sync.WaitGroup{}

	for _, conn := range connections {
		c := conn

		wg.Add(1)

		go func() {
			defer wg.Done()
			c.Close(ctx, websocket.StatusServiceRestart, "")
		}()
	}

	wg.Wait()

	return nil
}

// WithMaxFrameLimit sets the maximum frame size limit for incomming messages to the ConnectionRegistry.
// The limit parameter specifies the maximum frame size limit in bytes.
// This option can be used when creating a new ConnectionRegistry instance.
// The default frame size limit is 32768 bytes.
// If the limit is set to -1, the frame size limit is disabled.
// When the frame size limit is exceeded, the connection is closed with status 1009 (message too large).
func WithMaxFrameLimit(limit int64) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.frameSizeLimit = limit
	}
}

// WithConcurrencyLimit sets the maximum number of concurrent requests that can be handled by a connection.
// The default concurrency limit is 25.
// When the concurrency limit is exceeded, the connection stops reading messages until the number of concurrent requests decreases.
func WithConcurrencyLimit(limit uint) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.concurrencyLimit = limit
	}
}

package channel

import (
	"context"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

const (
	DefaultConcurencyLimitPerConnection = 25
	FrameSizeLimitInBytes               = 32768
)

// DefaultConnectionRegistry is default implementation of ConnectionRegistry
type DefaultConnectionRegistry struct {
	connections      map[string]wasabi.Connection
	onClose          chan string
	bufferPool       *bufferPool
	concurrencyLimit uint
	mu               sync.RWMutex
	frameSizeLimit   int64
}

type ConnectionRegistryOption func(*DefaultConnectionRegistry)

// NewDefaultConnectionRegistry creates new instance of DefaultConnectionRegistry
func NewDefaultConnectionRegistry(opts ...ConnectionRegistryOption) *DefaultConnectionRegistry {
	reg := &DefaultConnectionRegistry{
		connections:      make(map[string]wasabi.Connection),
		onClose:          make(chan string),
		concurrencyLimit: DefaultConcurencyLimitPerConnection,
		bufferPool:       newBufferPool(),
		frameSizeLimit:   FrameSizeLimitInBytes,
	}

	for _, opt := range opts {
		opt(reg)
	}

	go reg.handleClose()

	return reg
}

// AddConnection adds new Websocket connection to registry
func (r *DefaultConnectionRegistry) AddConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
) wasabi.Connection {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn := NewConnection(ctx, ws, cb, r.onClose, r.bufferPool, r.concurrencyLimit)
	r.connections[conn.ID()] = conn

	conn.ws.SetReadLimit(r.frameSizeLimit)

	return conn
}

// GetConnection returns connection by id
func (r *DefaultConnectionRegistry) GetConnection(id string) wasabi.Connection {
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

// WithMaxFrameLimit sets the maximum frame size limit for incomming messages to the ConnectionRegistry.
// The limit parameter specifies the maximum frame size limit in bytes.
// This option can be used when creating a new DefaultConnectionRegistry instance.
// The default frame size limit is 32768 bytes.
// If the limit is set to -1, the frame size limit is disabled.
// When the frame size limit is exceeded, the connection is closed with status 1009 (message too large).
func WithMaxFrameLimit(limit int64) ConnectionRegistryOption {
	return func(r *DefaultConnectionRegistry) {
		r.frameSizeLimit = limit
	}
}

package channel

import (
	"bytes"
	"context"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

const (
	DefaultConcurencyLimitPerConnection = 25
)

// DefaultConnectionRegistry is default implementation of ConnectionRegistry
type DefaultConnectionRegistry struct {
	connections      map[string]wasabi.Connection
	onClose          chan string
	bufferPool       *bufferPool
	concurrencyLimit uint
	mu               sync.RWMutex
}

// NewDefaultConnectionRegistry creates new instance of DefaultConnectionRegistry
func NewDefaultConnectionRegistry() *DefaultConnectionRegistry {
	reg := &DefaultConnectionRegistry{
		connections:      make(map[string]wasabi.Connection),
		onClose:          make(chan string),
		concurrencyLimit: DefaultConcurencyLimitPerConnection,
		bufferPool:       newBufferPool(),
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

type bufferPool struct {
	pool *sync.Pool
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (p *bufferPool) get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

func (p *bufferPool) put(b *bytes.Buffer) {
	b.Reset()
	p.pool.Put(b)
}

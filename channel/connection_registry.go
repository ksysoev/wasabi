package channel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

const (
	concurencyLimitPerConnection = 25
	frameSizeLimitInBytes        = 32768
	inActivityTimeout            = 0 * time.Second
	connectionLimt               = -1
)

type ConnectionHook func(wasabi.Connection)

// ConnectionRegistry is default implementation of ConnectionRegistry
type ConnectionRegistry struct {
	connections       map[string]wasabi.Connection
	onClose           chan string
	bufferPool        *bufferPool
	onConnect         ConnectionHook
	onDisconnect      ConnectionHook
	concurrencyLimit  uint
	connectionLimit   int
	frameSizeLimit    int64
	inActivityTimeout time.Duration
	mu                sync.RWMutex
	isClosed          bool
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
		connectionLimit:  connectionLimt,
	}

	for _, opt := range opts {
		opt(reg)
	}

	go reg.handleClose()

	return reg
}

// AddConnection adds new Websocket connection to registry
func (r *ConnectionRegistry) HandleConnection(
	ctx context.Context,
	ws *websocket.Conn,
	cb wasabi.OnMessage,
) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connectionLimit > 0 && len(r.connections) >= r.connectionLimit {
		ws.Close(websocket.StatusTryAgainLater, "Connection limit reached")
		return
	}

	if r.isClosed {
		ws.Close(websocket.StatusServiceRestart, "Server is shutting down")
		return
	}

	conn := NewConnection(ctx, ws, cb, r.onClose, r.bufferPool, r.concurrencyLimit, r.inActivityTimeout)
	r.connections[conn.ID()] = conn

	conn.ws.SetReadLimit(r.frameSizeLimit)

	if r.onConnect != nil {
		r.onConnect(conn)
	}

	conn.handleRequests()
}

// CanAccept checks if the connection registry can accept new connections.
// It returns true if the registry can accept new connections, and false otherwise.
func (r *ConnectionRegistry) CanAccept() bool {
	fmt.Println("Connection limit", r.connectionLimit)

	if r.connectionLimit <= 0 {
		return true
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	fmt.Println("Connections", len(r.connections))

	return len(r.connections) < r.connectionLimit
}

// GetConnection returns connection by id
func (r *ConnectionRegistry) GetConnection(id string) wasabi.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connections[id]
}

// handleClose handles connection cloasures and removes them from registry
func (r *ConnectionRegistry) handleClose() {
	wg := sync.WaitGroup{}

	for id := range r.onClose {
		r.mu.Lock()
		connection := r.connections[id]
		delete(r.connections, id)
		r.mu.Unlock()

		if r.onDisconnect != nil {
			wg.Add(1)

			go func() {
				defer wg.Done()
				r.onDisconnect(connection)
			}()
		}
	}

	wg.Wait()
}

// Shutdown closes all connections in the ConnectionRegistry.
// It sets the isClosed flag to true, indicating that the registry is closed.
// It then iterates over all connections, closes them with the given context,
// and waits for all closures to complete before returning.
func (r *ConnectionRegistry) Close(ctx ...context.Context) error {
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
			c.Close(websocket.StatusServiceRestart, "", ctx...)
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

// WithInActivityTimeout sets the inactivity timeout for the connection.
// The default inactivity timeout is 0 seconds, which means the timeout is disabled.
// When the inactivity timeout is enabled, the connection is closed if there are no messages received within the specified duration.
func WithInActivityTimeout(timeout time.Duration) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.inActivityTimeout = timeout
	}
}

// WithOnConnectHook sets the connection hook function that will be called when a new connection is established.
// The provided callback function `cb` will be invoked with the newly established connection as its argument.
// This function returns a ConnectionRegistryOption that can be used to configure a ConnectionRegistry.
func WithOnConnectHook(cb ConnectionHook) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.onConnect = cb
	}
}

// WithOnDisconnectHook sets the callback function to be executed when a connection is disconnected.
// The provided callback function should have the signature `func(connectionID string)`.
// It can be used to perform any necessary cleanup or logging operations when a connection is disconnected.
func WithOnDisconnectHook(cb ConnectionHook) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.onDisconnect = cb
	}
}

// WithConnectionLimit sets the maximum number of connections that can be accepted by the ConnectionRegistry.
// The default connection limit is -1, which means there is no limit on the number of connections.
// If the connection limit is set to a positive integer, the ConnectionRegistry will not accept new connections
// once the number of active connections reaches the specified limit.
func WithConnectionLimit(limit int) ConnectionRegistryOption {
	return func(r *ConnectionRegistry) {
		r.connectionLimit = limit
	}
}

package wasabi

import (
	"context"
	"net/http"

	"golang.org/x/net/websocket"
)

// Dispatcher is interface for dispatchers
type Dispatcher interface {
	Dispatch(conn Connection, data []byte)
}

// Middlewere is interface for middlewares
type Middlewere func(http.Handler) http.Handler

// DefaultChannel is default implementation of Channel
type DefaultChannel struct {
	path         string
	disptacher   Dispatcher
	connRegistry ConnectionRegistry
	ctx          context.Context
	middlewares  []Middlewere
}

// NewDefaultChannel creates new instance of DefaultChannel
// path - channel path
// dispatcher - dispatcher to use
// connRegistry - connection registry to use
// reqParser - request parser to use
// returns new instance of DefaultChannel
func NewDefaultChannel(
	path string,
	dispatcher Dispatcher,
	connRegistry ConnectionRegistry,
) *DefaultChannel {
	return &DefaultChannel{
		path:         path,
		disptacher:   dispatcher,
		connRegistry: connRegistry,
		middlewares:  make([]Middlewere, 0),
	}
}

// Path returns url path for channel
func (c *DefaultChannel) Path() string {
	return c.path
}

// Handler returns http.Handler for channel
func (c *DefaultChannel) Handler() http.Handler {
	var ctx context.Context

	saveCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = r.Context()
			next.ServeHTTP(w, r)
		})
	}

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		conn := c.connRegistry.AddConnection(ctx, ws, c.disptacher.Dispatch)
		conn.HandleRequests()
	})

	return c.setContext(c.wrapMiddleware(saveCtx(wsHandler)))
}

// SetContext sets context for channel
func (c *DefaultChannel) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// Use adds middlewere to channel
func (c *DefaultChannel) Use(middlewere Middlewere) {
	c.middlewares = append(c.middlewares, middlewere)
}

// useMiddleware applies middlewares to handler
func (c *DefaultChannel) wrapMiddleware(handler http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler
}

// setContext sets context for handler
func (c *DefaultChannel) setContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(c.ctx))
	})
}

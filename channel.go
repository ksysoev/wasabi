package wasabi

import (
	"context"
	"log/slog"
	"net/http"

	"golang.org/x/net/websocket"
)

// Channel is interface for channels
type Channel interface {
	Path() string
	SetContext(ctx context.Context)
	Handler() http.Handler
}

// Middlewere is interface for middlewares
type Middlewere func(http.Handler) http.Handler

// DefaultChannel is default implementation of Channel
type DefaultChannel struct {
	path         string
	disptacher   Dispatcher
	connRegistry ConnectionRegistry
	reqParser    RequestParser
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
	reqParser RequestParser,
) *DefaultChannel {
	return &DefaultChannel{
		path:         path,
		disptacher:   dispatcher,
		connRegistry: connRegistry,
		reqParser:    reqParser,
		middlewares:  make([]Middlewere, 0),
	}
}

// Path returns channel path
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
		conn := c.connRegistry.AddConnection(ctx, ws, c.onMessage)
		conn.HandleRequests()
	})

	return c.setContext(c.useMiddleware(saveCtx(wsHandler)))
}

// SetContext sets context for channel
func (c *DefaultChannel) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// Use adds middlewere to channel
func (c *DefaultChannel) Use(middlewere Middlewere) {
	c.middlewares = append(c.middlewares, middlewere)
}

// onMessage handles incoming messages
func handleRequestError(err error, conn Connection) {
	slog.Debug("Error parsing request: " + err.Error())
	resp := ResponseFromError(err)

	data, err := resp.String()
	if err != nil {
		slog.Debug("Error creating response: " + err.Error())
		return
	}

	conn.Send([]byte(data))
}

// useMiddleware applies middlewares to handler
func (c *DefaultChannel) useMiddleware(handler http.Handler) http.Handler {
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

// onMessage handles incoming messages
func (c *DefaultChannel) onMessage(conn Connection, data []byte) {
	req, err := c.reqParser.Parse(data)
	if err != nil {
		handleRequestError(err, conn)
	}

	req = req.WithContext(conn.Context())

	if err := c.disptacher.Dispatch(conn, req); err != nil {
		handleRequestError(err, conn)
	}
}

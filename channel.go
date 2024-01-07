package wasabi

import (
	"context"
	"log/slog"
	"net/http"

	"golang.org/x/net/websocket"
)

type Channel interface {
	Path() string
	SetContext(ctx context.Context)
	HTTPHandler() http.Handler
}

type Middlewere func(http.Handler) http.Handler

type DefaultChannel struct {
	path         string
	disptacher   Dispatcher
	connRegistry ConnectionRegistry
	reqParser    RequestParser
	ctx          context.Context
	middlewares  []Middlewere
}

func NewDefaultChannel(path string, dispatcher Dispatcher, connRegistry ConnectionRegistry, reqParser RequestParser) *DefaultChannel {
	return &DefaultChannel{
		path:         path,
		disptacher:   dispatcher,
		connRegistry: connRegistry,
		reqParser:    reqParser,
		middlewares:  make([]Middlewere, 0),
	}
}

func (c *DefaultChannel) Path() string {
	return c.path
}

func (c *DefaultChannel) HTTPHandler() http.Handler {
	var ctx context.Context
	saveCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = r.Context()
			next.ServeHTTP(w, r)
		})
	}

	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		conn := NewConnection(ctx, ws)
		conn.onMessageCB = func(conn Connection, data []byte) {
			req, err := c.reqParser.Parse(data)
			if err != nil {
				handleRequestError(err, conn)
			}

			req = req.WithContext(conn.Context())

			if err := c.disptacher.Dispatch(conn, req); err != nil {
				handleRequestError(err, conn)
			}
		}
		c.connRegistry.AddConnection(conn)

		conn.HandleRequest()
	})

	return c.setContext(c.useMiddleware(saveCtx(wsHandler)))
}

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

func (c *DefaultChannel) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *DefaultChannel) Use(middlewere Middlewere) {
	c.middlewares = append(c.middlewares, middlewere)
}

func (c *DefaultChannel) useMiddleware(handler http.Handler) http.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler
}

func (c *DefaultChannel) setContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(c.ctx))
	})
}

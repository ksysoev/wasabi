package wasabi

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type Channel interface {
	ConnectionHandler(ws *websocket.Conn)
	MiddlewareHandler(next http.Handler) http.Handler
	Path() string
}

type DefaultChannel struct {
	path         string
	disptacher   Dispatcher
	connRegistry ConnectionRegistry
}

func NewDefaultChannel(path string, dispatcher Dispatcher, connRegistry ConnectionRegistry) *DefaultChannel {
	return &DefaultChannel{
		path:         path,
		disptacher:   dispatcher,
		connRegistry: connRegistry,
	}
}

func (c *DefaultChannel) Path() string {
	return c.path
}

func (c *DefaultChannel) ConnectionHandler(ws *websocket.Conn) {
	conn := c.connRegistry.AddConnection(ws, c.disptacher.Dispatch)

	conn.HandleRequest()
}

func (c *DefaultChannel) MiddlewareHandler(next http.Handler) http.Handler {
	return next
}

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
	path       string
	disptacher Dispatcher
}

func NewDefaultChannel(path string, dispatcher Dispatcher) *DefaultChannel {
	return &DefaultChannel{path: path}
}

func (c *DefaultChannel) Path() string {
	return c.path
}

func (c *DefaultChannel) ConnectionHandler(ws *websocket.Conn) {
	conn := NewConnection(ws, c.disptacher.Dispatch)

	conn.HandleRequest()
}

func (c *DefaultChannel) MiddlewareHandler(next http.Handler) http.Handler {
	return next
}

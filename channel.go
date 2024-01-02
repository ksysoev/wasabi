package wasabi

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type Channel interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
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

func (c *DefaultChannel) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	conn := NewConnection()
	conn.Stash().Set("headers", r.Header)

	websocket.Handler(func(ws *websocket.Conn) {
		conn.ws = ws
		conn.onMessageCB = c.disptacher.Dispatch
		c.connRegistry.AddConnection(conn)

		conn.HandleRequest()
	}).ServeHTTP(w, r)
}

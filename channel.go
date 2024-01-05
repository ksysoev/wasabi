package wasabi

import (
	"context"
	"log/slog"
	"net/http"

	"golang.org/x/net/websocket"
)

type Channel interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Path() string
	SetContext(ctx context.Context)
}

type DefaultChannel struct {
	path         string
	disptacher   Dispatcher
	connRegistry ConnectionRegistry
	reqParser    RequestParser
	ctx          context.Context
}

func NewDefaultChannel(path string, dispatcher Dispatcher, connRegistry ConnectionRegistry, reqParser RequestParser) *DefaultChannel {
	return &DefaultChannel{
		path:         path,
		disptacher:   dispatcher,
		connRegistry: connRegistry,
		reqParser:    reqParser,
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
		conn.onMessageCB = func(conn *Connection, data []byte) {
			req, err := c.reqParser.Parse(data)
			if err != nil {
				handleRequestError(err, conn)
			}

			if err := c.disptacher.Dispatch(conn, req); err != nil {
				handleRequestError(err, conn)
			}
		}
		c.connRegistry.AddConnection(conn)

		conn.HandleRequest()
	}).ServeHTTP(w, r)
}

func handleRequestError(err error, conn *Connection) {
	slog.Debug("Error parsing request: " + err.Error())
	resp := ResponseFromError(err)

	data, err := resp.String()
	if err != nil {
		slog.Debug("Error creating response: " + err.Error())
		return
	}

	conn.SendResponse(data)
}

func (c *DefaultChannel) SetContext(ctx context.Context) {
	c.ctx = ctx
}

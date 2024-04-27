package dispatch

import (
	"log/slog"

	"github.com/ksysoev/wasabi"
)

// PipeDispatcher is a dispatcher that does not support any routing of requests
// but for single backend API gateways is enough
type PipeDispatcher struct {
	backend     wasabi.RequestHandler
	middlewares []RequestMiddlewere
}

// NewPipeDispatcher creates new instance of PipeDispatcher
func NewPipeDispatcher(backend wasabi.RequestHandler) *PipeDispatcher {
	return &PipeDispatcher{backend: backend}
}

// Dispatch dispatches request to backend
func (d *PipeDispatcher) Dispatch(conn wasabi.Connection, msgType wasabi.MessageType, data []byte) {
	req := NewRawRequest(conn.Context(), msgType, data)

	err := d.useMiddleware(d.backend).Handle(conn, req)
	if err != nil {
		slog.Error("Error handling request: " + err.Error())
	}
}

// Use adds middlewere to dispatcher
func (d *PipeDispatcher) Use(middlewere RequestMiddlewere) {
	d.middlewares = append(d.middlewares, middlewere)
}

// useMiddleware prepare request handler with middleweres chain
func (d *PipeDispatcher) useMiddleware(endpoint wasabi.RequestHandler) wasabi.RequestHandler {
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		endpoint = d.middlewares[i](endpoint)
	}

	return endpoint
}

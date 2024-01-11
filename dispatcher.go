package wasabi

import "log/slog"

// Dispatcher is interface for dispatchers
type Dispatcher interface {
	Dispatch(conn Connection, data []byte)
}

// PipeDispatcher is a dispatcher that does not support any routing of requests
// but for single backend API gateways is enough
type PipeDispatcher struct {
	backend     Backend
	reqParser   RequestParser
	middlewares []RequestMiddlewere
}

// RequestHandler is interface for request handlers
type RequestHandler interface {
	Handle(conn Connection, req Request) error
}

// RequestMiddlewere is interface for request middleweres
type RequestMiddlewere func(next RequestHandler) RequestHandler

// NewPipeDispatcher creates new instance of PipeDispatcher
func NewPipeDispatcher(backend Backend, reqParser RequestParser) *PipeDispatcher {
	return &PipeDispatcher{backend: backend, reqParser: reqParser}
}

// Dispatch dispatches request to backend
func (d *PipeDispatcher) Dispatch(conn Connection, data []byte) {
	req, err := d.reqParser.Parse(data)
	if err != nil {
		handleRequestError(err, conn)
	}

	req = req.WithContext(conn.Context())

	err = d.useMiddleware(d.backend).Handle(conn, req)
	if err != nil {
		handleRequestError(err, conn)
	}
}

// Use adds middlewere to dispatcher
func (d *PipeDispatcher) Use(middlewere RequestMiddlewere) {
	d.middlewares = append(d.middlewares, middlewere)
}

// useMiddleware prepare request handler with middleweres chain
func (d *PipeDispatcher) useMiddleware(endpoint RequestHandler) RequestHandler {
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		endpoint = d.middlewares[i](endpoint)
	}

	return endpoint
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

	err = conn.Send([]byte(data))
	if err != nil {
		slog.Debug("Error sending response: " + err.Error())
		return
	}
}

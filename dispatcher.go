package wasabi

// Dispatcher is interface for dispatchers
type Dispatcher interface {
	Dispatch(conn Connection, req Request) error
}

// PipeDispatcher is a dispatcher that does not support any routing of requests
// but for single backend API gateways is enough
type PipeDispatcher struct {
	backend     Backend
	middlewares []RequestMiddlewere
}

// RequestHandler is interface for request handlers
type RequestHandler interface {
	Handle(conn Connection, req Request) error
}

// RequestMiddlewere is interface for request middleweres
type RequestMiddlewere func(next RequestHandler) RequestHandler

// NewPipeDispatcher creates new instance of PipeDispatcher
func NewPipeDispatcher(backend Backend) *PipeDispatcher {
	return &PipeDispatcher{backend: backend}
}

// Dispatch dispatches request to backend
func (d *PipeDispatcher) Dispatch(conn Connection, req Request) error {
	return d.useMiddleware(d.backend).Handle(conn, req)
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

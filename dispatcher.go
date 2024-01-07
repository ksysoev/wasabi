package wasabi

type Dispatcher interface {
	Dispatch(conn *Connection, req Request) error
}

type PipeDispatcher struct {
	backend     Backend
	middlewares []RequestMiddlewere
}

type RequestHandler interface {
	Handle(conn *Connection, req Request) error
}

type RequestMiddlewere func(next RequestHandler) RequestHandler

func NewPipeDispatcher(backend Backend) *PipeDispatcher {
	return &PipeDispatcher{backend: backend}
}

func (d *PipeDispatcher) Dispatch(conn *Connection, req Request) error {
	return d.useMiddleware(d.backend).Handle(conn, req)
}

func (d *PipeDispatcher) Use(middlewere RequestMiddlewere) {
	d.middlewares = append(d.middlewares, middlewere)
}

func (d *PipeDispatcher) useMiddleware(endpoint RequestHandler) RequestHandler {
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		endpoint = d.middlewares[i](endpoint)
	}

	return endpoint
}

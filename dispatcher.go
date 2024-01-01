package wasabi

type Dispatcher interface {
	Dispatch(conn *Connection, req Request) error
}

type PipeDispatcher struct {
	backend Backend
}

func NewPipeDispatcher(backend Backend) *PipeDispatcher {
	return &PipeDispatcher{backend: backend}
}

func (d *PipeDispatcher) Dispatch(conn *Connection, req Request) error {
	return d.backend.Forward(conn, req)
}

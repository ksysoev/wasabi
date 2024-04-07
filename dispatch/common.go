package dispatch

import "github.com/ksysoev/wasabi"

// RequestMiddlewere is interface for request middleweres
type RequestMiddlewere func(next RequestHandler) RequestHandler

// RequestHandlerFunc is a function that implements RequestHandler interface
type RequestHandlerFunc func(conn wasabi.Connection, req wasabi.Request) error

// Handle implements RequestHandler interface
func (f RequestHandlerFunc) Handle(conn wasabi.Connection, req wasabi.Request) error {
	return f(conn, req)
}

// TODO: Shall we moove it wasabi package?
// RequestHandler is interface for request handlers
type RequestHandler interface {
	Handle(conn wasabi.Connection, req wasabi.Request) error
}

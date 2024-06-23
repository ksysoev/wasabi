package dispatch

import (
	"context"

	"github.com/ksysoev/wasabi"
)

// RequestMiddlewere is interface for request middleweres
type RequestMiddlewere func(next wasabi.RequestHandler) wasabi.RequestHandler

// RequestHandlerFunc is a function that implements RequestHandler interface
type RequestHandlerFunc func(conn wasabi.Connection, req wasabi.Request) error

// Handle implements RequestHandler interface
func (f RequestHandlerFunc) Handle(conn wasabi.Connection, req wasabi.Request) error {
	return f(conn, req)
}

type RequestParser func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request

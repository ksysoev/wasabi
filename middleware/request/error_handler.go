package request

import (
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type ErrorHandler func(conn wasabi.Connection, req wasabi.Request, err error) error

// NewErrorHandlingMiddleware returns a new error handling middleware that wraps the provided request handler.
// The middleware calls the provided `onError` function when an error occurs during request handling.
// It returns a new request handler that executes the provided `next` request handler and handles any errors that occur.
func NewErrorHandlingMiddleware(onError ErrorHandler) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			err := next.Handle(conn, req)

			if err != nil {
				return onError(conn, req, err)
			}

			return nil
		})
	}
}

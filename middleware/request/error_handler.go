package request

import (
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type ErrorHandler func(conn wasabi.Connection, req wasabi.Request, err error) error

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

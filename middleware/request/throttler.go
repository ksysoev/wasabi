package request

import (
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type token struct{}

func NewTrottlerMiddleware(limit uint) func(next dispatch.RequestHandler) dispatch.RequestHandler {
	sem := make(chan token, limit)

	return func(next dispatch.RequestHandler) dispatch.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			select {
			case sem <- token{}:
				defer func() { <-sem }()
				return next.Handle(conn, req)
			case <-conn.Context().Done():
				return conn.Context().Err()
			}
		})
	}
}

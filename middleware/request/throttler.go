package request

import (
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type token struct{}

// NewTrottlerMiddleware creates a new throttler middleware that limits the number of concurrent requests.
// The `limit` parameter specifies the maximum number of concurrent requests allowed.
// It returns a function that takes a `wasabi.RequestHandler` as input and returns a new `wasabi.RequestHandler`
// that enforces the throttling limit.
func NewTrottlerMiddleware(limit uint) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	sem := make(chan token, limit)

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			select {
			case sem <- token{}:
				defer func() { <-sem }()
				return next.Handle(conn, req)
			case <-req.Context().Done():
				return req.Context().Err()
			}
		})
	}
}

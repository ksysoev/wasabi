package request

import (
	"context"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewSetTimeoutMiddleware returns a middleware that sets a timeout for each request.
// The timeout duration is specified by the 'timeout' parameter.
// The returned middleware is a function that takes a 'next' request handler as input
// and returns a new request handler that applies the timeout to the incoming request.
func NewSetTimeoutMiddleware(timeout time.Duration) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			ctx := req.Context()
			ctx, cancel := context.WithTimeout(ctx, timeout)

			defer cancel()

			req = req.WithContext(ctx)

			return next.Handle(conn, req)
		})
	}
}

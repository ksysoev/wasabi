package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewRetryMiddleware returns a new retry middleware that wraps the provided `next` request handler.
// The middleware retries the request a maximum of `maxRetries` times with a delay of `interval` between each retry.
// If the request succeeds at any retry, the middleware returns `nil`. If all retries fail, it returns the last error encountered.
func NewRetryMiddleware(maxRetries int, interval time.Duration, retryPolicy func(maxRetries int, interval time.Duration, next wasabi.RequestHandler, conn wasabi.Connection, req wasabi.Request) error) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			return retryPolicy(maxRetries, interval, next, conn, req)
		})
	}
}

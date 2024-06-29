package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewRetryMiddleware returns a new retry middleware that wraps the provided `next` request handler.
// The middleware accepts a `RetryConfig` consisting of:
//   - `maxRetries`: maximum number of retries allowed
//   - `GetRetryInterval`: A higher order function that returns a function to calculate next retry interval
//
// If the request succeeds at any retry, the middleware returns `nil`. If all retries fail, it returns the last error encountered.
func NewRetryMiddleware(getRetryInterval GetRetryInterval, shouldRetry ShouldRetry) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			var err error

			ticker := time.NewTicker(getRetryInterval(0))

			defer ticker.Stop()

			for {
				condition, i := shouldRetry()

				if !condition {
					break
				}

				err = next.Handle(conn, req)
				if err == nil {
					return nil
				}

				ticker.Reset(getRetryInterval(i))

				select {
				case <-req.Context().Done():
					return req.Context().Err()
				case <-ticker.C:
				}
			}

			return err
		})
	}
}

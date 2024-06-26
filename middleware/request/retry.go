package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewRetryMiddleware returns a new retry middleware that wraps the provided `next` request handler.
// The middleware retries the request a maximum of `maxRetries` times with a delay of `interval` between each retry.
// If the request succeeds at any retry, the middleware returns `nil`. If all retries fail, it returns the last error encountered.
func NewRetryMiddleware(retryConfig RetryConfig) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			var err error

			ticker := time.NewTicker(retryConfig.seedInterval)

			defer ticker.Stop()

			for i := 0; i < retryConfig.maxRetries; i++ {
				err = next.Handle(conn, req)
				if err == nil {
					return nil
				}

				ticker.Reset(GetRetryInterval(retryConfig.retryPolicy, retryConfig.seedInterval, i, retryConfig.delayFactor))

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

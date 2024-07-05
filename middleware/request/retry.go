package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewRetryMiddleware returns a new retry middleware that wraps the provided `next` request handler.
// The middleware accepts 3 params consisting of:
//   - `getRetryInterval`: A higher order function that returns a function to calculate next retry interval
//   - `maxRetries`: maximum number of retries allowed for this middleware
//   - `predicates`: optional function(s) that determine whether the encountered
//     errors needs retry. If multiple functions are provided, it will be
//     executed in sequence and stop the moment a function returns false
//
// If the request succeeds at any retry, the middleware returns `nil`. If all retries fail, it returns the last error encountered.
func NewRetryMiddleware(getRetryInterval GetRetryInterval, maxRetries int, predicates ...func(e error) bool) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			var err error

			ticker := time.NewTicker(getRetryInterval(0))

			defer ticker.Stop()

			for i := 0; i < maxRetries; i++ {
				err = next.Handle(conn, req)
				if err == nil {
					return nil
				}

				for _, pred := range predicates {
					ok := pred(err)
					if !ok {
						return err
					}
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

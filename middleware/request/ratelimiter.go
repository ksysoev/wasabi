package request

import (
	"time"

	"github.com/ksysoev/ratestor"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewRateLimiterMiddleware returns a middleware function that implements rate limiting for incoming requests.
// The `requestLimit` function is used to determine the rate limit for each request based on the provided request object.
// The `requestLimit` function takes a `wasabi.Request` object as input and returns the rate limit key, period, and limit.
// The `stor` variable is an instance of `ratestor.RateStor` used to store and manage rate limit information.
// The returned middleware function takes a `wasabi.RequestHandler` as input and returns a new `wasabi.RequestHandler`
// that performs rate limiting before passing the request to the next handler in the chain.
func NewRateLimiterMiddleware(requestLimit func(wasabi.Request) (key string, period time.Duration, limit uint64)) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	stor := ratestor.NewRateStor()

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			key, period, limit := requestLimit(req)
			if err := stor.Allow(key, period, limit); err != nil {
				return err
			}

			return next.Handle(conn, req)
		})
	}
}

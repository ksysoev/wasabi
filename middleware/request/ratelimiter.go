package request

import (
	"time"

	"github.com/ksysoev/ratestor"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

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

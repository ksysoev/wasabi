package request

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"golang.org/x/sync/singleflight"
)

type ResponseCache struct {
	data    []byte
	msgType wasabi.MessageType
}

// NewCacheMiddleware returns a new cache middleware that wraps the provided `next` request handler.
// The middleware caches the response of the request for a given `duration` and returns the cached response if the request is made again within the cache duration.
// If the request is made after the cache duration, the middleware forwards the request to the next handler and caches the response.
func NewCacheMiddleware(requestCache func(r wasabi.Request) (cacheKey string, ttl time.Duration)) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		cache, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,
			MaxCost:     1 << 30,
			BufferItems: 64,
		})
		if err != nil {
			//Should not happen, error is due to invalid configuration only
			panic(err)
		}

		group := &singleflight.Group{}

		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			cacheKey, ttl := requestCache(req)

			if cacheKey == "" {
				return next.Handle(conn, req)
			}

			resp, err, _ := group.Do(cacheKey, func() (interface{}, error) {
				if item, found := cache.Get(cacheKey); found {
					return item, nil
				}

				var resp ResponseCache

				connWrapper := channel.NewConnectionWrapper(conn, channel.WithSendWrapper(func(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
					resp.data = msg
					resp.msgType = msgType
					return nil
				}))

				err := next.Handle(connWrapper, req)

				if err != nil {
					return nil, err
				}

				cache.SetWithTTL(cacheKey, resp, 0, ttl)

				return resp, nil
			})

			if err != nil {
				return err
			}

			respCache := resp.(ResponseCache)

			conn.Send(respCache.msgType, respCache.data)

			return nil
		})
	}
}

package request

import (
	"fmt"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"golang.org/x/sync/singleflight"
)

type responseCache struct {
	data    []byte
	msgType wasabi.MessageType
}

// NewCacheMiddleware returns a new cache middleware that wraps the provided `next` request handler.
// The middleware caches the response of the request for a given `duration` and returns the cached response if the request is made again within the cache duration.
// If the request is made after the cache duration, the middleware forwards the request to the next handler and caches the response.
func NewCacheMiddleware(requestCache func(r wasabi.Request) (cacheKey string, ttl time.Duration)) (middleware func(next wasabi.RequestHandler) wasabi.RequestHandler, cacheCloser func()) {
	cache := ttlcache.New[string, responseCache]()

	done := make(chan struct{})
	go func() {
		cache.Start()
		close(done)
	}()

	closer := func() {
		cache.Stop()
		<-done
	}

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		group := &singleflight.Group{}

		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			cacheKey, ttl := requestCache(req)

			if cacheKey == "" {
				return next.Handle(conn, req)
			}

			resp, err, _ := group.Do(cacheKey, func() (interface{}, error) {
				if item := cache.Get(cacheKey); item != nil {
					return item.Value(), nil
				}

				var resp responseCache

				connWrapper := channel.NewConnectionWrapper(conn, channel.WithSendWrapper(func(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
					resp.data = msg
					resp.msgType = msgType
					return nil
				}))

				err := next.Handle(connWrapper, req)

				if err != nil {
					return nil, err
				}
				if ttl > 0 {
					cache.Set(cacheKey, resp, ttl)
				}

				return resp, nil
			})

			if err != nil {
				return err
			}

			respCache, ok := resp.(responseCache)
			if !ok {
				return fmt.Errorf("invalid response cache type")
			}

			return conn.Send(respCache.msgType, respCache.data)
		})
	}, closer
}

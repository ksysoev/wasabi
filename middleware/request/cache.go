package request

import (
	"fmt"
	"sync"
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

// NewCacheMiddleware returns a middleware function that implements caching for requests.
// This middleware also implement debouncing pattern for requests to avoid duplicate requests hitting backend.
//
// Parameters:
// - requestCache: A function that takes a `wasabi.Request` and returns a cache key and TTL duration.
//
// Returns:
// - middleware: A function that takes a `wasabi.RequestHandler` and returns a new `wasabi.RequestHandler` with caching functionality.
// - cacheCloser: A function that stops the cache and performs cleanup.
func NewCacheMiddleware(requestCache func(r wasabi.Request) (cacheKey string, ttl time.Duration)) (middleware func(next wasabi.RequestHandler) wasabi.RequestHandler, cacheCloser func()) {
	cache := ttlcache.New[string, responseCache]()
	started := make(chan struct{})
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()

		close(started)
		cache.Start()
	}()

	<-started

	closer := func() {
		cache.Stop()
		wg.Wait()
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

				connWrapper := channel.NewConnectionWrapper(conn, channel.WithSendWrapper(func(_ wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
					resp.data = msg
					resp.msgType = msgType

					return nil
				}))

				if err := next.Handle(connWrapper, req); err != nil {
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

			if req.Context().Err() != nil {
				return req.Context().Err()
			}

			respCache, ok := resp.(responseCache)
			if !ok {
				return fmt.Errorf("invalid response cache type")
			}

			return conn.Send(respCache.msgType, respCache.data)
		})
	}, closer
}

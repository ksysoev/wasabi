package http

import (
	"context"
	"net/http"
	"sync"
)

// NewStashMiddleware returns a middleware function that adds a stash to the request context.
// The stash is a synchronized map that can be used to store and retrieve values during the connection lifecycle.
// The middleware function takes the next http.Handler as input and returns a new http.Handler that wraps the next handler.
// The returned handler adds the stash to the request context and then calls the next handler to process the request.
func NewStashMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, Stash, &sync.Map{})
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetStash retrieves the stash from the provided context.
// It takes a single parameter ctx of type context.Context.
// It returns a pointer to a sync.Map representing the stash if found, otherwise nil.
func GetStash(ctx context.Context) *sync.Map {
	if stash, ok := ctx.Value(Stash).(*sync.Map); ok {
		return stash
	}

	return nil
}

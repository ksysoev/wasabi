package http

import (
	"context"
	"net/http"
	"sync"
)

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

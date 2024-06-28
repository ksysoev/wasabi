package http

import (
	"net/http"
)

// NewProtectedMiddleware returns a middleware function that performs verification of the request.
// It takes a `verifyToken` function as an argument. This function is used to verify the token
// The token string should be set as as `Authorization` header in the http request
// If the token is valid, it calls the next handler in the chain.
// If the token is invalid or missing, it sends an HTTP 401 Unauthorized response.
func NewProtectedMiddleware(verifyToken func(token string) error) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				Unauthorized(w, "Missing authorization header in request", func(w http.ResponseWriter) { w.WriteHeader(http.StatusUnauthorized) })
				return
			}

			tokenString = tokenString[len("Bearer "):]

			err := verifyToken(tokenString)
			if err != nil {
				Unauthorized(w, "Invalid token", func(w http.ResponseWriter) { w.WriteHeader(http.StatusUnauthorized) })
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

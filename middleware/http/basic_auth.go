package http

import "net/http"

// NewBasicAuthMiddleware returns a middleware function that performs basic authentication.
// It takes a map of users and passwords, and a realm string as input.
// The returned middleware function checks if the request contains valid basic authentication credentials.
// If the credentials are valid, it calls the next handler in the chain.
// If the credentials are invalid or missing, it sends an HTTP 401 Unauthorized response.
func NewBasicAuthMiddleware(users map[string]string, realm string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			if !ok {
				unauthorized(w, realm)
				return
			}

			if p, ok := users[user]; !ok || p != pass {
				unauthorized(w, realm)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// unauthorized sends an HTTP 401 Unauthorized response with the specified realm.
func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

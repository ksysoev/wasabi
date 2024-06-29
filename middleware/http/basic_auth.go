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
				unauthorized(w, "Unauthorized", setRealm(realm))
				return
			}

			if p, ok := users[user]; !ok || p != pass {
				unauthorized(w, "Unauthorized", setRealm(realm))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// sets the specified realm in response header
func setRealm(realm string) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	}
}

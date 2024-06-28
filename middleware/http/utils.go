package http

import "net/http"

// Unauthorized sends an HTTP 401 Unauthorized response with the specified operations made on responseWriter.
func Unauthorized(w http.ResponseWriter, message string, writerPostOperations ...func(w http.ResponseWriter)) {
	for _, operation := range writerPostOperations {
		operation(w)
	}

	http.Error(w, message, http.StatusUnauthorized)
}

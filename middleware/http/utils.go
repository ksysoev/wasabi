package http

import "net/http"

// unauthorized sends an HTTP 401 unauthorized response with the specified operations made on responseWriter.
func unauthorized(w http.ResponseWriter, message string, writerPostOperations ...func(w http.ResponseWriter)) {
	for _, operation := range writerPostOperations {
		operation(w)
	}

	http.Error(w, message, http.StatusUnauthorized)
}

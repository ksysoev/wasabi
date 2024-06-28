package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProtectedAuthMiddleware(t *testing.T) {
	verify := func(token string) error {
		if token != "SECRET" {
			return errors.New("invalid token")
		}

		return nil
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewProtectedMiddleware(verify)
	handler := middleware(nextHandler)

	// Test case 1: token verification successful
	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	req1.Header.Set("Authorization", "Bearer SECRET")

	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)

	resp1 := w1.Result()

	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Test case 2: Invalid credentials
	req2, _ := http.NewRequest("GET", "/", http.NoBody)
	req2.Header.Set("Authorization", "Bearer Idon'tKnow")

	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w2, req2)

	resp2 := w2.Result()

	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, resp2.StatusCode)
	}

	// Test case 3: Missing credentials
	req3, _ := http.NewRequest("GET", "/", http.NoBody)
	w3 := httptest.NewRecorder()

	handler.ServeHTTP(w3, req3)

	resp3 := w3.Result()

	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, resp3.StatusCode)
	}
}

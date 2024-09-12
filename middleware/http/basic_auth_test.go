package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewBasicAuthMiddleware(t *testing.T) {
	users := map[string]string{
		"admin": "password",
		"user":  "123456",
	}
	realm := "Test Realm"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewBasicAuthMiddleware(users, realm)
	handler := middleware(nextHandler)

	// Test case 1: Valid credentials
	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	req1.SetBasicAuth("admin", "password")

	w1 := httptest.NewRecorder()

	handler.ServeHTTP(w1, req1)

	resp1 := w1.Result()

	defer resp1.Body.Close()

	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp1.StatusCode)
	}

	// Test case 2: Invalid credentials
	req2, _ := http.NewRequest("GET", "/", http.NoBody)
	req2.SetBasicAuth("admin", "wrongpassword")

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

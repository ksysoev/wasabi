package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()

	Unauthorized(w, "Test Error Message", func(w http.ResponseWriter) {
		w.Header().Set("TEST-HEADER", "TEST-HEADER-VALUE")
	})

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	testHeaderValue := resp.Header.Get("TEST-HEADER")

	if testHeaderValue != "TEST-HEADER-VALUE" {
		t.Errorf("Expected TEST-HEADER value to be %q but got %q", "TEST-HEADER-VALUE", testHeaderValue)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err == nil {
		message := string(bytes)
		if strings.EqualFold(message, "Test Error Message") {
			t.Errorf("Expected error message to be %s but got %s", "Test Error Message", message)
		}
	} else {
		t.Errorf("Error while reading response body - %v", err)
	}
}

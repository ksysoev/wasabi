package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNewStashMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test if the stash value is set in the request context
		stash := r.Context().Value(Stash)
		if stash == nil {
			t.Error("Expected stash value to be set in the request context")
		}

		// Test if the stash value is of type *sync.Map
		_, ok := stash.(*sync.Map)
		if !ok {
			t.Error("Expected stash value to be of type *sync.Map")
		}

		// Test if the next handler is called
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewStashMiddleware()
	testServer := httptest.NewServer(middleware(handler))

	defer testServer.Close()

	// Send a test request to the server
	resp, err := http.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request to test server: %v", err)
	}
	defer resp.Body.Close()

	// Test if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetStash(t *testing.T) {
	ctx := context.Background()
	stash := &sync.Map{}
	ctx = context.WithValue(ctx, Stash, stash)

	retrievedStash := GetStash(ctx)
	if retrievedStash == nil {
		t.Errorf("Expected stash to be retrieved from context, but got nil")
	}

	if retrievedStash != stash {
		t.Errorf("Expected retrieved stash to be %v, but got %v", stash, retrievedStash)
	}
}

func TestGetStash_NotFound(t *testing.T) {
	ctx := context.Background()

	retrievedStash := GetStash(ctx)
	if retrievedStash != nil {
		t.Errorf("Expected nil when stash is not found in context, but got %v", retrievedStash)
	}
}

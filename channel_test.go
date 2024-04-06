package wasabi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDefaultChannel(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	if channel.path != path {
		t.Errorf("Unexpected path: got %q, expected %q", channel.path, path)
	}

	if channel.disptacher != dispatcher {
		t.Errorf("Unexpected dispatcher: got %v, expected %v", channel.disptacher, dispatcher)
	}

	if channel.connRegistry != connRegistry {
		t.Errorf("Unexpected connection registry: got %v, expected %v", channel.connRegistry, connRegistry)
	}

	if len(channel.middlewares) != 0 {
		t.Errorf("Unexpected number of middlewares: got %d, expected %d", len(channel.middlewares), 0)
	}
}
func TestDefaultChannel_Path(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	if channel.Path() != path {
		t.Errorf("Unexpected path: got %q, expected %q", channel.Path(), path)
	}
}
func TestDefaultChannel_Handler(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)
	channel.SetContext(context.Background())

	// Call the Handler method
	handler := channel.Handler()

	if handler == nil {
		t.Errorf("Unexpected nil handler")
	}
}
func TestDefaultChannel_SetContext(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	ctx := context.Background()
	channel.SetContext(ctx)

	if channel.ctx != ctx {
		t.Errorf("Unexpected context: got %v, expected %v", channel.ctx, ctx)
	}
}
func TestDefaultChannel_Use(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Custom middleware logic
		})
	}

	channel.Use(middleware)

	if len(channel.middlewares) != 1 {
		t.Errorf("Unexpected number of middlewares: got %d, expected %d", len(channel.middlewares), 1)
	}
}

func TestDefaultChannel_wrapMiddleware(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	// Create a mock handler
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock handler logic
	})

	// Create mock middlewares
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock middleware 1 logic
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock middleware 2 logic
		})
	}

	// Add the mock middlewares to the channel
	channel.middlewares = append(channel.middlewares, middleware1, middleware2)

	// Call the wrapMiddleware method
	wrappedHandler := channel.wrapMiddleware(mockHandler)

	// Assert that the wrappedHandler is the result of applying the middlewares to the mockHandler
	if wrappedHandler == nil {
		t.Errorf("Unexpected nil wrappedHandler")
	}
}
func TestDefaultChannel_SetContextMiddleware(t *testing.T) {
	path := "/test/path"
	dispatcher := NewMockDispatcher(t)
	connRegistry := NewMockConnectionRegistry(t)

	channel := NewDefaultChannel(path, dispatcher, connRegistry)

	// Create a mock handler
	var ctx context.Context

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
	})

	// Create a mock request
	mockRequest := httptest.NewRequest(http.MethodGet, "/", http.NoBody)

	// Create a mock response recorder
	mockResponseRecorder := httptest.NewRecorder()

	// Set the context for the channel
	channel.SetContext(context.WithValue(context.Background(), struct{ key string }{"test"}, "test"))

	// Wrap the mock handler with the setContext middleware
	wrappedHandler := channel.setContext(mockHandler)

	// Call the wrappedHandler with the mock request and response recorder
	wrappedHandler.ServeHTTP(mockResponseRecorder, mockRequest)

	// Assert that the context of the request is set to the channel's context
	if ctx != channel.ctx {
		t.Errorf("Unexpected context: got %v, expected %v", ctx, channel.ctx)
	}
}

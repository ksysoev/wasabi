package server

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ksysoev/wasabi/mocks"
)

type testCtxKey string

func TestNewServer(t *testing.T) {
	addr := ":8080"
	server := NewServer(addr)

	if server.addr != addr {
		t.Errorf("Expected port %s, but got %s", addr, server.addr)
	}

	if len(server.channels) != 0 {
		t.Errorf("Expected empty channels slice, but got %d channels", len(server.channels))
	}

	if server.mutex == nil {
		t.Error("Expected non-nil mutex")
	}

	server = NewServer("")

	if server.addr != ":http" {
		t.Errorf("Expected default port :http, but got %s", server.addr)
	}
}
func TestServer_AddChannel(t *testing.T) {
	// Create a new Server instance
	server := NewServer(":0")

	// Create a new channel
	channel := mocks.NewMockChannel(t)
	channel.EXPECT().Path().Return("test")

	// Add the channel to the server
	server.AddChannel(channel)

	// Check if the channel was added correctly
	if len(server.channels) != 1 {
		t.Errorf("Expected 1 channel, but got %d channels", len(server.channels))
	}

	if server.channels[0].Path() != "test" {
		t.Errorf("Expected channel name 'test', but got '%s'", server.channels[0].Path())
	}
}

func TestServer_WithBaseContext(t *testing.T) {
	// Create a new Server instance with a base context
	ctx := context.WithValue(context.Background(), testCtxKey("test"), "test")

	server := NewServer(":0", WithBaseContext(ctx))

	// Check if the base context was set correctly
	if server.baseCtx == nil {
		t.Error("Expected non-nil base context")
	}

	if server.baseCtx.Value(testCtxKey("test")) != "test" {
		t.Errorf("Expected context value 'test', but got '%s'", server.baseCtx.Value("test"))
	}
}

func TestServer_WithReadinessChan(t *testing.T) {
	// Create a new Server instance with a base context
	ready := make(chan struct{})
	server := NewServer(":0", WithReadinessChan(ready))

	if server.ready == nil {
		t.Error("Expected non-nil channel")
	}

	close(server.ready)
	_, ok := <-ready
	if ok {
		t.Error("Expected closed channel")
	}
}

func TestServer_Run(t *testing.T) {
	noOfReruns := []int{0, 1, 2}

	for _, run := range noOfReruns {
		t.Run(fmt.Sprintf("%d times of calling Run", run), func(t *testing.T) {
			// Create a new Server instance
			ready := make(chan struct{})
			server := NewServer(":0", WithReadinessChan(ready))

			channel := mocks.NewMockChannel(t)
			channel.EXPECT().Path().Return("/test")
			channel.EXPECT().Handler().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

			server.AddChannel(channel)

			// for signaling server stopped without error
			done := make(chan struct{})

			// Run the server
			go func() {
				err := server.Run()
				switch err {
				case nil:
					close(done)
				default:
					t.Errorf("Got unexpected error: %v", err)
				}
			}()

			// Wait for server to be ready
			select {
			case <-ready:
			case <-time.After(1 * time.Second):
				t.Error("Expected server to start")
			}

			// Test that calling Run on a running server returns
			// ErrServerAlreadyRunning
			for i := 0; i < run; i++ {
				if err := server.Run(); err != ErrServerAlreadyRunning {
					t.Error("Should return ErrServerAlreadyRunning when triggered run on running server")
				}
			}

			if err := server.handler.Close(); err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}

			select {
			case <-done:
			case <-time.After(1 * time.Second):
				t.Error("Expected server to stop")
			}
		})
	}
}

func TestServer_Close(t *testing.T) {
	// Create a new Server instance
	ready := make(chan struct{})
	server := NewServer(":0", WithReadinessChan(ready))

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Create a mock channel
	channel := mocks.NewMockChannel(t)
	channel.EXPECT().Path().Return("/test")
	channel.EXPECT().Handler().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	channel.EXPECT().Close(ctx).Return(nil)

	server.AddChannel(channel)

	defer cancel()

	// Start the server in a separate goroutine
	done := make(chan struct{})

	// Run the server
	go func() {
		err := server.Run()
		switch err {
		case nil:
			close(done)
		default:
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
	}

	// Call the Shutdown method
	err := server.Close(ctx)
	if err != nil {
		t.Errorf("Unexpected error shutting down server: %v", err)
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to stop")
	}
}

func TestServer_Close_NoContext(t *testing.T) {
	// Create a new Server instance
	ready := make(chan struct{})
	server := NewServer(":0", WithReadinessChan(ready))

	// Create a mock channel
	channel := mocks.NewMockChannel(t)
	channel.EXPECT().Path().Return("/test")
	channel.EXPECT().Handler().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	channel.EXPECT().Close().Return(nil)

	server.AddChannel(channel)

	// Start the server in a separate goroutine
	done := make(chan struct{})

	// Run the server
	go func() {
		err := server.Run()
		switch err {
		case nil:
			close(done)
		default:
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to start")
	}

	// Call the Shutdown method
	err := server.Close()
	if err != nil {
		t.Errorf("Unexpected error shutting down server: %v", err)
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to stop")
	}
}

func TestServer_Addr(t *testing.T) {
	// Create a new Server instance
	done := make(chan struct{})

	server := NewServer(":0", WithReadinessChan(done))
	defer server.Close()

	// Create a mock channel
	channel := mocks.NewMockChannel(t)
	channel.EXPECT().Path().Return("/test")
	channel.EXPECT().Handler().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	channel.EXPECT().Close().Return(nil)

	server.AddChannel(channel)

	if server.Addr() != nil {
		t.Error("Expected nil address for server that is not running")
	}

	// Start the server in a separate goroutine
	go func() {
		err := server.Run()
		if err != nil {
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	// Wait for the server to fully start
	<-done

	addr := server.Addr()

	if addr == nil {
		t.Error("Expected non-empty address")
	}
}

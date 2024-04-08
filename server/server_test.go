package server

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi/mocks"
)

func TestServer_Shutdown_with_context(t *testing.T) {
	// Create a new Server instance
	server := NewServer(0)

	// Create a new context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	doneChan := make(chan struct{})
	// Start the server in a separate goroutine
	go func() {
		_ = server.Run(ctx)

		close(doneChan)
	}()

	cancel()

	select {
	case <-doneChan:
	case <-time.After(1 * time.Second):
		t.Error("Server did not stop")
	}
}

func TestNewServer(t *testing.T) {
	port := uint16(8080)
	server := NewServer(port)

	if server.port != port {
		t.Errorf("Expected port %d, but got %d", port, server.port)
	}

	if len(server.channels) != 0 {
		t.Errorf("Expected empty channels slice, but got %d channels", len(server.channels))
	}

	if server.mutex == nil {
		t.Error("Expected non-nil mutex")
	}
}
func TestServer_AddChannel(t *testing.T) {
	// Create a new Server instance
	server := NewServer(0)

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

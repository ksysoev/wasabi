package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "net/http/pprof" //nolint:gosec // pprof is used for testing profile endpoint

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

func TestServer_WithConfig(t *testing.T) {
	// Create a new Server instance with config
	serverConfig := Config{
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       10 * time.Second,
	}

	server := NewServer(":0", WithServerConfig(serverConfig))

	// Check if the base context was set correctly
	if server.baseCtx == nil {
		t.Error("Expected non-nil base context")
	}

	if server.GetServerConfig() == DefaultConfig {
		t.Errorf("Expected non-Default context value for server config")
	}

	config := server.GetServerConfig()

	if config.ReadHeaderTimeout != 1*time.Second {
		t.Errorf("Expected config ReadHeaderTimeout to be %s but got %s", 1*time.Second, config.ReadHeaderTimeout)
	}

	if config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected config ReadTimeout to be %s but got %s", 10*time.Second, config.ReadTimeout)
	}
}

func TestServer_WithDefaultConfig(t *testing.T) {
	// Create a new vanilla Server instance
	server := NewServer(":0")

	// Check if the base context was set correctly
	if server.baseCtx == nil {
		t.Error("Expected non-nil base context")
	}

	if server.GetServerConfig() != DefaultConfig {
		t.Errorf("Expected Default context value for server config")
	}
}

func TestServer_WithBaseContextAndConfig(t *testing.T) {
	serverConfig := Config {
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout: 20 * time.Second,
	}
	ctx := context.WithValue(context.Background(), testCtxKey("test"), "test")

	// Create 2 servers with different order of optional methods
	server1 := NewServer(":0", WithBaseContext(ctx), WithServerConfig(serverConfig))
	server2 := NewServer(":1", WithServerConfig(serverConfig), WithBaseContext(ctx))

	// Check if the base context was set correctly
	if server1.baseCtx == nil || server2.baseCtx == nil {
		t.Error("Expected non-nil base contexts for both servers")
	}

	if server1.baseCtx.Value(testCtxKey("test")) != "test" {
		t.Errorf("Expected context value 'test', but got '%s'", server1.baseCtx.Value("test"))
	}

	if server2.baseCtx.Value(testCtxKey("test")) != "test" {
		t.Errorf("Expected context value 'test', but got '%s'", server2.baseCtx.Value("test"))
	}

	if (server1.GetServerConfig() != serverConfig) {
		t.Errorf("Expected config for server1 to be %s but got %s", serverConfig, server1.GetServerConfig())
	}

	if (server2.GetServerConfig() != serverConfig) {
		t.Errorf("Expected config for server2 to be %s but got %s", serverConfig, server2.GetServerConfig())
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

	if _, ok := <-ready; ok {
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
func TestServer_WithTLS(t *testing.T) {
	// Create a new Server instance
	server := NewServer(":0")
	// Set TLS configuration using WithTLS
	certPath := "/path/to/cert.pem"
	keyPath := "/path/to/key.pem"

	// #nosec G402 - InsecureSkipVerify is used for testing purposes
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	WithTLS(certPath, keyPath, tlsConfig)(server)

	// Check if the certificate and key paths are set correctly
	if server.certPath != certPath {
		t.Errorf("Expected certificate path %s, but got %s", certPath, server.certPath)
	}

	if server.keyPath != keyPath {
		t.Errorf("Expected key path %s, but got %s", keyPath, server.keyPath)
	}

	// Check if the TLS configuration is set correctly
	if server.handler.TLSConfig == nil {
		t.Error("Expected non-nil TLS configuration")
	}

	if server.handler.TLSConfig.InsecureSkipVerify != true {
		t.Error("Expected InsecureSkipVerify to be true")
	}

	err := server.Run()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Got unexpected error: %v", err)
	}
}

func TestServer_WithProfilerEndpoint(t *testing.T) {
	ready := make(chan struct{})
	// Create a new Server instance
	server := NewServer(":0", WithReadinessChan(ready))

	// Check if the profiler endpoint is disabled by default
	if server.pprofEnabled {
		t.Error("Expected profiler endpoint to be disabled")
	}

	// Apply the WithProfilerEndpoint option
	WithProfilerEndpoint()(server)

	// Check if the profiler endpoint is enabled
	if !server.pprofEnabled {
		t.Error("Expected profiler endpoint to be enabled")
	}

	go func() {
		err := server.Run()
		if err != nil {
			t.Errorf("Got unexpected error: %v", err)
		}
	}()

	defer server.Close()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Expected server to start")
	}

	// Check if the profiler endpoint is enabled
	resp, err := http.Get("http://" + server.Addr().String() + "/debug/pprof/")

	if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", resp.StatusCode)
	}
}

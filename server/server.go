// Package server provides a HTTP server with custom timeouts and error handling.
//
// This package is designed to be used with the wasabi package for handling HTTP requests.
// It provides a Server struct that wraps the standard http.Server with additional functionality.
//
// The Server struct includes a base context, a net.Listener for accepting connections, and a mutex for thread safety.
// It also includes a ready channel that can be used to signal when the server is ready to accept connections.
//
// The server uses custom read header and read timeouts, defined as constants.
// These timeouts help to prevent slow client attacks by limiting the amount of time the server will wait for a client to send its request.
//
// Usage:
//
//	s := server.NewServer(":8080")
//	s.AddChannel(channel.NewChannel("/path", handler))
//	err := s.Start()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// This will start a new server on port 8080 with the provided handler.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/ksysoev/wasabi"
	"golang.org/x/exp/slog"
)

var ErrServerAlreadyRunning = fmt.Errorf("server is already running")

type Server struct {
	certPath     string
	keyPath      string
	baseCtx      context.Context
	listener     net.Listener
	listenerLock *sync.Mutex
	mutex        *sync.Mutex
	handler      *http.Server
	ready        chan<- struct{}
	addr         string
	channels     []wasabi.Channel
	pprofEnabled bool
}

type Option func(*Server)

type ctxConfigKey struct{}

// WithReadinessChan sets ch to [Server] and will be closed once the [Server] is
// ready to accept connection. Typically used in testing after calling [Run]
// method and waiting for ch to close, before continuing with test logics.
func WithReadinessChan(ch chan<- struct{}) Option {
	return func(s *Server) {
		s.ready = ch
	}
}

// NewServer creates new instance of Wasabi server
// port - port to listen on
// returns new instance of Server
func NewServer(addr string, opts ...Option) *Server {
	if addr == "" {
		addr = ":http"
	}

	server := &Server{
		addr:         addr,
		channels:     make([]wasabi.Channel, 0, 1),
		mutex:        &sync.Mutex{},
		listenerLock: &sync.Mutex{},
		baseCtx:      context.Background(),
	}

	for _, opt := range opts {
		opt(server)
	}

	serverConfig := server.GetServerConfig()

	server.handler = &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: serverConfig.ReadHeaderTimeout,
		ReadTimeout:       serverConfig.ReadTimeout,
		BaseContext: func(_ net.Listener) context.Context {
			return server.baseCtx
		},
	}

	return server
}

// Helper to fetch server config. Returns Default config if base ctx has no config set
func (s *Server) GetServerConfig() Config {
	config, ok := s.baseCtx.Value(ctxConfigKey{}).(Config)
	if !ok {
		return DefaultConfig
	}

	return config
}

// AddChannel adds new channel to server
func (s *Server) AddChannel(channel wasabi.Channel) {
	s.channels = append(s.channels, channel)
}

// Run starts the server
// returns error if server is already running
// or if server fails to start
func (s *Server) Run() (err error) {
	if !s.mutex.TryLock() {
		return ErrServerAlreadyRunning
	}

	defer s.mutex.Unlock()

	mux := http.NewServeMux()

	for _, channel := range s.channels {
		mux.Handle(
			channel.Path(),
			channel.Handler(),
		)
	}

	if s.pprofEnabled {
		slog.Info("Profiler endpoint enabled on /debug/pprof/")
		mux.Handle("/debug/pprof/", http.DefaultServeMux)
	}

	s.handler.Handler = mux

	s.listenerLock.Lock()
	s.listener, err = net.Listen("tcp", s.addr)
	s.listenerLock.Unlock()

	if err != nil {
		return err
	}

	slog.Info("Starting app server on " + s.listener.Addr().String())

	// Signals that server can accept connections
	if s.ready != nil {
		close(s.ready)
	}

	if s.certPath != "" && s.keyPath != "" {
		err = s.handler.ServeTLS(s.listener, s.certPath, s.keyPath)
	} else {
		err = s.handler.Serve(s.listener)
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server and all its channels.
// It waits for all channels to be shut down before returning.
// If the context is canceled before all channels are shut down, it returns the context error.
// If any error occurs during the shutdown process, it returns the first error encountered.
func (s *Server) Close(ctx ...context.Context) error {
	done := make(chan error)

	go func() {
		defer close(done)

		if len(ctx) > 0 {
			done <- s.handler.Shutdown(ctx[0])
			return
		}

		done <- s.handler.Close()
	}()

	wg := sync.WaitGroup{}

	for _, channel := range s.channels {
		c := channel

		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := c.Close(ctx...); err != nil {
				slog.Error("Error shutting down channel:" + err.Error())
			}
		}()
	}

	wg.Wait()

	return <-done
}

// Addr returns the server's network address.
// If the server is not running, it returns nil.
func (s *Server) Addr() net.Addr {
	s.listenerLock.Lock()
	defer s.listenerLock.Unlock()

	if s.listener == nil {
		return nil
	}

	return s.listener.Addr()
}

// BaseContext optionally specifies based context that will be used for all connections.
// If not specified, context.Background() will be used.
func WithBaseContext(ctx context.Context) Option {
	if ctx == nil {
		panic("nil context")
	}

	return func(s *Server) {
		config := s.GetServerConfig()
		s.baseCtx = context.WithValue(ctx, ctxConfigKey{}, config)
	}
}

// WithTLS is an option function that configures the server to use TLS (Transport Layer Security).
// It sets the certificate and key file paths, and optionally allows custom TLS configuration.
// The certificate and key file paths must be provided as arguments.
// If a custom TLS configuration is provided, it will be applied to the server's handler.
func WithTLS(certFile, keyFile string, config ...*tls.Config) Option {
	return func(s *Server) {
		s.certPath = certFile
		s.keyPath = keyFile

		if len(config) > 0 {
			s.handler.TLSConfig = config[0]
		}
	}
}

// WithProfilerEndpoint is an option function that enables the profiler endpoint for the server.
// Enabling the profiler endpoint allows profiling and performance monitoring of the server.
// The profiler endpoint is available at /debug/pprof/.
// To use the profiler endpoint, import the net/http/pprof package in your application.
// Example:
//
//	import _ "net/http/pprof"
//
// The profiler endpoint is disabled by default.
func WithProfilerEndpoint() Option {
	return func(s *Server) {
		s.pprofEnabled = true
	}
}

// WithServerConfig is an option function that overrides the default configuration settings of the server.
// This can give the client more control over the server feature functions
func WithServerConfig(config Config) Option {
	return func(s *Server) {
		s.baseCtx = context.WithValue(s.baseCtx, ctxConfigKey{}, config)
	}
}

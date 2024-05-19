package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ksysoev/wasabi"
	"golang.org/x/exp/slog"
)

const (
	ReadHeaderTimeout = 3 * time.Second
	ReadTimeout       = 30 * time.Second
)

var ErrServerAlreadyRunning = fmt.Errorf("server is already running")

type Server struct {
	baseCtx  context.Context
	listener net.Listener
	mutex    *sync.Mutex
	handler  *http.Server
	addr     string
	channels []wasabi.Channel
}

type Option func(*Server)

// NewServer creates new instance of Wasabi server
// port - port to listen on
// returns new instance of Server
func NewServer(addr string, opts ...Option) *Server {
	if addr == "" {
		addr = ":http"
	}

	server := &Server{
		addr:     addr,
		channels: make([]wasabi.Channel, 0, 1),
		mutex:    &sync.Mutex{},
		baseCtx:  context.Background(),
	}

	for _, opt := range opts {
		opt(server)
	}

	server.handler = &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		BaseContext: func(_ net.Listener) context.Context {
			return server.baseCtx
		},
	}

	return server
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

	s.handler.Handler = mux

	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	slog.Info("Starting app server on " + s.listener.Addr().String())

	err = s.handler.Serve(s.listener)

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
		s.baseCtx = ctx
	}
}

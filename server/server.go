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

type Server struct {
	mutex    *sync.Mutex
	channels []wasabi.Channel
	addr     string
	http     *http.Server
	baseCtx  context.Context
}

type ServerOption func(*Server)

// NewServer creates new instance of Wasabi server
// port - port to listen on
// returns new instance of Server
func NewServer(addr string, opts ...ServerOption) *Server {
	server := &Server{
		addr:     addr,
		channels: make([]wasabi.Channel, 0, 1),
		mutex:    &sync.Mutex{},
		baseCtx:  context.Background(),
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

// AddChannel adds new channel to server
func (s *Server) AddChannel(channel wasabi.Channel) {
	s.channels = append(s.channels, channel)
}

// Run starts server
// ctx - context
// returns error if any
func (s *Server) Run() error {
	if !s.mutex.TryLock() {
		return fmt.Errorf("server is already running")
	}

	defer s.mutex.Unlock()

	mux := http.NewServeMux()

	for _, channel := range s.channels {
		mux.Handle(
			channel.Path(),
			channel.Handler(),
		)
	}

	slog.Info("Starting app server on " + s.addr)

	s.http = &http.Server{
		Addr:              s.addr,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		Handler:           mux,
		BaseContext: func(_ net.Listener) context.Context {
			return s.baseCtx
		},
	}

	return s.http.ListenAndServe()
}

// BaseContext optionally specifies based context that will be used for all connections.
// If not specified, context.Background() will be used.
func WithBaseContext(ctx context.Context) ServerOption {
	if ctx == nil {
		panic("nil context")
	}

	return func(s *Server) {
		s.baseCtx = ctx
	}
}

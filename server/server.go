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
func (s *Server) Run() error {
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

	slog.Info("Starting app server on " + s.addr)

	return s.handler.ListenAndServe()
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

package wasabi

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/exp/slog"
)

// Channel is interface for channels
type Channel interface {
	Path() string
	SetContext(ctx context.Context)
	Handler() http.Handler
}

const (
	ReadHeaderTimeout = 3 * time.Second
	ReadTimeout       = 30 * time.Second
)

type Server struct {
	channels []Channel
	port     uint16
}

// NewServer creates new instance of Wasabi server
// port - port to listen on
// returns new instance of Server
func NewServer(port uint16) *Server {
	return &Server{
		port:     port,
		channels: make([]Channel, 0, 1),
	}
}

// AddChannel adds new channel to server
func (s *Server) AddChannel(channel Channel) {
	s.channels = append(s.channels, channel)
}

// Run starts server
// ctx - context
// returns error if any
func (s *Server) Run(ctx context.Context) error {
	listen := ":" + strconv.Itoa(int(s.port))

	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()

	for _, channel := range s.channels {
		channel.SetContext(execCtx)
		mux.Handle(
			channel.Path(),
			channel.Handler(),
		)
	}

	slog.Info("Starting app server on " + listen)

	server := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		Handler:           mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

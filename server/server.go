package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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
	mutex     *sync.Mutex
	channels  []wasabi.Channel
	port      uint16
	isRunning bool
}

// NewServer creates new instance of Wasabi server
// port - port to listen on
// returns new instance of Server
func NewServer(port uint16) *Server {
	return &Server{
		port:     port,
		channels: make([]wasabi.Channel, 0, 1),
		mutex:    &sync.Mutex{},
	}
}

// AddChannel adds new channel to server
func (s *Server) AddChannel(channel wasabi.Channel) {
	s.channels = append(s.channels, channel)
}

// Run starts server
// ctx - context
// returns error if any
func (s *Server) Run(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

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

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		<-execCtx.Done()

		slog.Info("Shutting down app server on " + listen)

		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown app server", "error", err)
		}

		wg.Done()
	}()

	err := server.ListenAndServe()

	cancel()

	wg.Wait()

	if err != http.ErrServerClosed {
		return err
	}

	return nil
}

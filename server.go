package wasabi

import (
	"context"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"
)

type Server struct {
	port     uint16
	channels []Channel
}

func NewServer(port uint16) *Server {

	return &Server{
		port:     port,
		channels: make([]Channel, 0, 1),
	}
}

func (s *Server) AddChannel(channel Channel) {
	s.channels = append(s.channels, channel)
}

func (s *Server) Run(ctx context.Context) error {
	listen := ":" + strconv.Itoa(int(s.port))

	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()

	for _, channel := range s.channels {
		channel.SetContext(execCtx)
		mux.Handle(
			channel.Path(),
			channel.HTTPHandler(),
		)
	}

	slog.Info("Starting app server on " + listen)

	err := http.ListenAndServe(listen, mux)
	if err != nil {
		return err
	}

	return nil
}

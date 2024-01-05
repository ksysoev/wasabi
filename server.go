package wasabi

import (
	"context"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"
)

type Server struct {
	port uint16
	mux  *http.ServeMux
	ctx  context.Context
}

func NewServer(port uint16) *Server {
	return &Server{
		port: port,
		mux:  http.NewServeMux(),
	}
}

func (s *Server) AddChannel(channel Channel) {
	s.mux.Handle(
		channel.Path(),
		channel,
	)
}

func (s *Server) Run(ctx context.Context) error {
	listen := ":" + strconv.Itoa(int(s.port))

	execCtx, cancel := context.WithCancel(ctx)
	s.ctx = execCtx
	defer cancel()

	slog.Info("Starting app server on " + listen)

	err := http.ListenAndServe(listen, s.mux)
	if err != nil {
		return err
	}

	return nil
}

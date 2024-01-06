package wasabi

import (
	"context"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"
)

type Server struct {
	port      uint16
	mux       *http.ServeMux
	ctx       context.Context
	cancelCtx context.CancelFunc
}

func NewServer(port uint16, ctx context.Context) *Server {
	execCtx, cancel := context.WithCancel(ctx)

	return &Server{
		port:      port,
		mux:       http.NewServeMux(),
		ctx:       execCtx,
		cancelCtx: cancel,
	}
}

func (s *Server) AddChannel(channel Channel) {
	s.mux.Handle(
		channel.Path(),
		channel,
	)

	channel.SetContext(s.ctx)
}

func (s *Server) Run() error {
	listen := ":" + strconv.Itoa(int(s.port))

	defer s.cancelCtx()

	slog.Info("Starting app server on " + listen)

	err := http.ListenAndServe(listen, s.mux)
	if err != nil {
		return err
	}

	return nil
}

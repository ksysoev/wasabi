package wasabi

import (
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"
)

type Server struct {
	port uint16
	mux  *http.ServeMux
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

func (s *Server) Run() error {
	listen := ":" + strconv.Itoa(int(s.port))

	slog.Info("Starting app server on " + listen)

	err := http.ListenAndServe(listen, s.mux)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

	return nil
}

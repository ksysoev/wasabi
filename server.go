package wasabi

import (
	"io"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

type Server struct {
	port uint16
}

func NewServer(port uint16) *Server {
	return &Server{port: port}
}

func (s *Server) Run() error {
	listen := ":" + strconv.Itoa(int(s.port))
	http.Handle("/", websocket.Handler(EchoServer))

	slog.Info("Starting app server on " + listen)

	err := http.ListenAndServe(listen, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}

	return nil
}

func EchoServer(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

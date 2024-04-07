package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/middleware/request"
)

const (
	Port = 8080
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := wasabi.NewBackend(func(req wasabi.Request) (*http.Request, error) {
		bodyReader := bytes.NewBufferString(string(req.Data()))
		httpReq, err := http.NewRequest("GET", "http://localhost:8081/", bodyReader)
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")

		return httpReq, nil
	})

	ErrHandler := request.NewErrorHandlingMiddleware(func(conn wasabi.Connection, req wasabi.Request, err error) error {

		if conn.Context().Err() != nil {
			return nil
		}

		if req.Context().Err() == nil {
			slog.Error("Error handling request", "error", err)
		}

		conn.Send([]byte("Failed to process request: " + err.Error()))
		return nil
	})

	connRegistry := wasabi.NewDefaultConnectionRegistry()
	dispatcher := wasabi.NewPipeDispatcher(backend)
	dispatcher.Use(ErrHandler)
	dispatcher.Use(request.NewTrottlerMiddleware(10))

	server := wasabi.NewServer(Port)
	channel := wasabi.NewDefaultChannel("/", dispatcher, connRegistry)

	server.AddChannel(channel)

	if err := server.Run(context.Background()); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

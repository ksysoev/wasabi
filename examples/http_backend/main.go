package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/middleware/request"
	"github.com/ksysoev/wasabi/server"
)

const (
	Port = 8080
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := backend.NewBackend(func(req wasabi.Request) (*http.Request, error) {
		httpReq, err := http.NewRequest("GET", "http://localhost:8081/", bytes.NewBuffer(req.Data()))
		if err != nil {
			return nil, err
		}

		return httpReq, nil
	})

	ErrHandler := request.NewErrorHandlingMiddleware(func(conn wasabi.Connection, req wasabi.Request, err error) error {
		conn.Send([]byte("Failed to process request: " + err.Error()))
		return nil
	})

	dispatcher := dispatch.NewPipeDispatcher(backend)
	dispatcher.Use(ErrHandler)
	dispatcher.Use(request.NewTrottlerMiddleware(10))

	channel := channel.NewDefaultChannel("/", dispatcher)

	server := server.NewServer(Port)
	server.AddChannel(channel)

	if err := server.Run(context.Background()); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ksysoev/wasabi"
)

const (
	Port = 8080
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := wasabi.NewBackend(func(req wasabi.Request) (*http.Request, error) {
		bodyReader := bytes.NewBufferString(string(req.Data()))
		httpReq, err := http.NewRequest("GET", "http://localhost:8080/", bodyReader)
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Content-Type", "application/json")

		return httpReq, nil
	})

	connRegistry := wasabi.NewDefaultConnectionRegistry()
	dispatcher := wasabi.NewPipeDispatcher(backend)
	server := wasabi.NewServer(Port)
	channel := wasabi.NewDefaultChannel("/", dispatcher, connRegistry)

	server.AddChannel(channel)

	if err := server.Run(context.Background()); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

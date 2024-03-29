package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ksysoev/wasabi"
)

const (
	Port = 8080
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := wasabi.NewBackend("http://localhost:8081")
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

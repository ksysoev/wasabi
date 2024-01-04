package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ksysoev/wasabi"
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")
	backend := wasabi.NewBackend("http://localhost:8081")

	connRegistry := wasabi.NewDefaultConnectionRegistry()
	dispatcher := wasabi.NewPipeDispatcher(backend)

	server := wasabi.NewServer(8080)
	channel := wasabi.NewDefaultChannel("/", dispatcher, connRegistry, &wasabi.JSONRPCRequestParser{})
	server.AddChannel(channel)

	err := server.Run()

	if err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

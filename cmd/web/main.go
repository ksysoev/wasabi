package main

import (
	"log/slog"
	"os"

	"github.com/ksysoev/wasabi"
)

func main() {

	backend := wasabi.NewBackend("http://localhost:8081")
	dispatcher := wasabi.NewPipeDispatcher(backend)

	server := wasabi.NewServer(8080, dispatcher)
	err := server.Run()

	if err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

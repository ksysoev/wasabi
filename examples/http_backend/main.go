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
	"github.com/ksysoev/wasabi/server"
)

const (
	Addr = ":8080"
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	be := backend.NewBackend(func(req wasabi.Request) (*http.Request, error) {
		httpReq, err := http.NewRequest("GET", "http://localhost:8081/", bytes.NewBuffer(req.Data()))
		if err != nil {
			return nil, err
		}

		return httpReq, nil
	})

	dispatcher := dispatch.NewRouterDispatcher(be, func(_ wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
	})

	ch := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	serv := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	serv.AddChannel(ch)

	if err := serv.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

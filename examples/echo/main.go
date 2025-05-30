package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/server"
)

const (
	Addr = ":8080"
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return conn.Send(wasabi.MsgTypeText, req.Data())
	})

	dispatcher := dispatch.NewRouterDispatcher(backend, func(_ wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
	})
	ch := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	serv := server.NewServer(Addr, server.WithBaseContext(context.Background()), server.WithProfilerEndpoint())

	serv.AddChannel(ch)

	if err := serv.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}

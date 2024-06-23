package main

import (
	"context"
	"fmt"
	"log/slog"
	_ "net/http/pprof"
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

	dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
	})
	channel := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	server := server.NewServer(Addr, server.DefaultServerConfig, server.WithBaseContext(context.Background()), server.WithProfilerEndpoint())

	server.AddChannel(channel)

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}

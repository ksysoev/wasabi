package main

import (
	"context"
	"fmt"
	"log/slog"
	_ "net/http/pprof"
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

	backend := backend.NewWSBackend(
		"wss://ws.derivws.com/websockets/v3?app_id=1089",
		func(r wasabi.Request) (wasabi.MessageType, []byte, error) {
			switch r.RoutingKey() {
			case "text":
				return wasabi.MsgTypeText, r.Data(), nil
			case "binary":
				return wasabi.MsgTypeBinary, r.Data(), nil
			default:
				var t wasabi.MessageType
				return t, nil, fmt.Errorf("unsupported request type: %s", r.RoutingKey())
			}
		},
	)

	dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(conn.Context(), msgType, data)
	})
	channel := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	server := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	server.AddChannel(channel)

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}

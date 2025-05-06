package main

import (
	"context"
	"fmt"
	"log/slog"
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

	be := backend.NewWSBackend(
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

	dispatcher := dispatch.NewRouterDispatcher(be, func(conn wasabi.Connection, _ context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(conn.Context(), msgType, data)
	})
	ch := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	serv := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	serv.AddChannel(ch)

	if err := serv.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	fmt.Println("Server is stopped")
	os.Exit(0)
}

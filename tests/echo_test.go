// Package tests contains integration tests for wasabi library.
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/server"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TestEcho(t *testing.T) {
	// Message to be sent to server and expected to be received
	expected := "a message for server"

	backend := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return conn.Send(wasabi.MsgTypeText, req.Data())
	})

	dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
	})
	ch := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))

	ready := make(chan struct{})
	s := server.NewServer(":0", server.DefaultServerConfig, server.WithBaseContext(context.Background()), server.WithReadinessChan(ready))
	s.AddChannel(ch)

	go func() {
		if err := s.Run(); err != nil {
			t.Error("Fail to start app server", "error", err)
		}
	}()

	select {
	case <-ready:
	case <-time.After(1 * time.Second):
		t.Error("Server is expected to start")
	}

	url := "ws://" + s.Addr().String()

	ws, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Errorf("Unexpected error dialing websocket: %s", err)
	}

	defer ws.Close(websocket.CloseStatus(nil), "exiting")

	err = wsjson.Write(context.Background(), ws, expected)
	if err != nil {
		t.Errorf("Unexpected error writing to websocket: %s", err)
	}

	var actual any

	if err := wsjson.Read(context.Background(), ws, &actual); err != nil {
		t.Errorf("Unexpected error reading from websocket: %s", err)
	}

	if expected != actual {
		t.Errorf("Expected: %s, actual: %s", expected, actual)
	}
}

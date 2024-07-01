package wasabi

import (
	"context"
	"net/http"

	"nhooyr.io/websocket"
)

type MessageType = websocket.MessageType

const (
	MsgTypeText   MessageType = websocket.MessageText
	MsgTypeBinary MessageType = websocket.MessageBinary
)

type Request interface {
	Data() []byte
	RoutingKey() string
	Context() context.Context
	WithContext(ctx context.Context) Request
}

// Dispatcher is interface for dispatchers
type Dispatcher interface {
	Dispatch(conn Connection, msgType MessageType, data []byte)
}

// OnMessage is type for OnMessage callback
type OnMessage func(conn Connection, msgType MessageType, data []byte)

// Connection is interface for connections
type Connection interface {
	Send(msgType MessageType, msg []byte) error
	Context() context.Context
	ID() string
	Close(status websocket.StatusCode, reason string, closingCtx ...context.Context) error
}

// RequestHandler is interface for request handlers
type RequestHandler interface {
	Handle(conn Connection, req Request) error
}

// Channel is interface for channels
type Channel interface {
	Path() string
	Handler() http.Handler
	Close(ctx ...context.Context) error
}

// ConnectionRegistry is interface for connection registries
type ConnectionRegistry interface {
	HandleConnection(
		ctx context.Context,
		ws *websocket.Conn,
		cb OnMessage,
	)
	GetConnection(id string) Connection
	Close(ctx ...context.Context) error
	CanAccept() bool
}

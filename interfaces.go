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

type Backend interface {
	Handle(conn Connection, r Request) error
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
	HandleRequests()
	Close(closingCtx context.Context, status websocket.StatusCode, reason string) error
}

// RequestHandler is interface for request handlers
type RequestHandler interface {
	Handle(conn Connection, req Request) error
}

// Channel is interface for channels
type Channel interface {
	Path() string
	SetContext(ctx context.Context)
	Handler() http.Handler
}

// ConnectionRegistry is interface for connection registries
type ConnectionRegistry interface {
	AddConnection(
		ctx context.Context,
		ws *websocket.Conn,
		cb OnMessage,
	) Connection
	GetConnection(id string) Connection
}

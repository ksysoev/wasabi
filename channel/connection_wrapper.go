package channel

import (
	"context"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type WrapperOptions func(*ConnectionWrapper)

type SendWrapper func(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error
type CloseWrapper func(conn wasabi.Connection, status websocket.StatusCode, reason string, closingCtx ...context.Context) error

type ConnectionWrapper struct {
	connection     wasabi.Connection
	onSendWrapper  SendWrapper
	onCloseWrapper CloseWrapper
}

func NewConnectionWrapper(connection wasabi.Connection, opts ...WrapperOptions) *ConnectionWrapper {
	wrapper := &ConnectionWrapper{
		connection: connection,
	}

	for _, opt := range opts {
		opt(wrapper)
	}

	return wrapper
}

func (cw *ConnectionWrapper) ID() string {
	return cw.connection.ID()
}

func (cw *ConnectionWrapper) Context() context.Context {
	return cw.connection.Context()
}

func (cw *ConnectionWrapper) HandleRequests() {
	cw.connection.HandleRequests()
}

func (cw *ConnectionWrapper) Send(msgType wasabi.MessageType, msg []byte) error {
	if cw.onSendWrapper != nil {
		return cw.onSendWrapper(cw.connection, msgType, msg)
	}

	return cw.connection.Send(msgType, msg)
}

func (cw *ConnectionWrapper) Close(status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	if cw.onCloseWrapper != nil {
		return cw.onCloseWrapper(cw.connection, status, reason, closingCtx...)
	}

	return cw.connection.Close(status, reason, closingCtx...)
}

func WithSendWrapper(wrapper SendWrapper) WrapperOptions {
	return func(cw *ConnectionWrapper) {
		cw.onSendWrapper = wrapper
	}
}

func WithCloseWrapper(wrapper CloseWrapper) WrapperOptions {
	return func(cw *ConnectionWrapper) {
		cw.onCloseWrapper = wrapper
	}
}

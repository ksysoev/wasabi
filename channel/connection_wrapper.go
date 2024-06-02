package channel

import (
	"context"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

// WrapperOptions is a function type that represents options for configuring a ConnectionWrapper.
// It is used to modify the behavior of the ConnectionWrapper by applying various options.
type WrapperOptions func(*ConnectionWrapper)

// SendWrapper is a function type that wraps the Send method of a ConnectionWrapper.
type SendWrapper func(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error

// CloseWrapper is a function type that wraps the Close method of a ConnectionWrapper.
type CloseWrapper func(conn wasabi.Connection, status websocket.StatusCode, reason string, closingCtx ...context.Context) error

// ConnectionWrapper is a wrapper around a wasabi.Connection that allows for custom behavior to be applied to the connection.
type ConnectionWrapper struct {
	connection     wasabi.Connection
	onSendWrapper  SendWrapper
	onCloseWrapper CloseWrapper
}

// NewConnectionWrapper creates a new ConnectionWrapper instance with the given connection and options.
// It applies the provided WrapperOptions to the wrapper before returning it.
func NewConnectionWrapper(connection wasabi.Connection, opts ...WrapperOptions) *ConnectionWrapper {
	wrapper := &ConnectionWrapper{
		connection: connection,
	}

	for _, opt := range opts {
		opt(wrapper)
	}

	return wrapper
}

// ID returns the ID of the ConnectionWrapper.
func (cw *ConnectionWrapper) ID() string {
	return cw.connection.ID()
}

// Context returns the context associated with the connection wrapper.
func (cw *ConnectionWrapper) Context() context.Context {
	return cw.connection.Context()
}

// Send sends a message of the specified type and content over the connection.
// If an onSendWrapper function is set, it will be called instead of directly sending the message.
// The onSendWrapper function should have the signature func(connection Connection, msgType MessageType, msg []byte) error.
// If there is no onSendWrapper function set, the message will be sent directly using the underlying connection.
// Returns an error if there was a problem sending the message.
func (cw *ConnectionWrapper) Send(msgType wasabi.MessageType, msg []byte) error {
	if cw.onSendWrapper != nil {
		return cw.onSendWrapper(cw.connection, msgType, msg)
	}

	return cw.connection.Send(msgType, msg)
}

// Close closes the connection with the specified status code and reason.
// It also accepts an optional closing context.
// If an onCloseWrapper function is set, it will be called instead of directly closing the connection.
// The onCloseWrapper function should have the same signature as the Connection.Close method.
func (cw *ConnectionWrapper) Close(status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
	if cw.onCloseWrapper != nil {
		return cw.onCloseWrapper(cw.connection, status, reason, closingCtx...)
	}

	return cw.connection.Close(status, reason, closingCtx...)
}

// WithSendWrapper returns a WrapperOptions function that sets the onSendWrapper
// field of the ConnectionWrapper to the provided wrapper.
func WithSendWrapper(wrapper SendWrapper) WrapperOptions {
	return func(cw *ConnectionWrapper) {
		cw.onSendWrapper = wrapper
	}
}

// WithCloseWrapper returns a WrapperOptions function that sets the onCloseWrapper
// field of the ConnectionWrapper to the provided wrapper.
func WithCloseWrapper(wrapper CloseWrapper) WrapperOptions {
	return func(cw *ConnectionWrapper) {
		cw.onCloseWrapper = wrapper
	}
}

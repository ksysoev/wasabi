package dispatch

import (
	"context"
	"fmt"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewRouterDispatcher(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)

	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return mocks.NewMockRequest(t)
	}

	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	if dispatcher.defaultBackend != defaultBackend {
		t.Errorf("Expected defaultBackend to be %v, but got %v", defaultBackend, dispatcher.defaultBackend)
	}

	if len(dispatcher.backendMap) != 0 {
		t.Errorf("Expected backendMap to be empty, but got %v", dispatcher.backendMap)
	}
}
func TestRouterDispatcher_AddBackend(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return mocks.NewMockRequest(t)
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	backend := mocks.NewMockBackend(t)
	routingKeys := []string{"key1", "key2"}

	err := dispatcher.AddBackend(backend, routingKeys)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	for _, key := range routingKeys {
		if dispatcher.backendMap[key] != backend {
			t.Errorf("Expected backend %v for routing key %s, but got %v", backend, key, dispatcher.backendMap[key])
		}
	}

	// Test adding backend with existing routing key
	err = dispatcher.AddBackend(backend, routingKeys)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		expectedErr := fmt.Errorf("backend for routing key %s already exists", routingKeys[0])
		if err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, but got %v", expectedErr, err)
		}
	}
}

func TestRouterDispatcher_DispatchDefaultBackend(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)

	req := mocks.NewMockRequest(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return req
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().Context().Return(context.Background())

	data := []byte("test data")

	// Test case 1: Request with existing routing key
	routingKey := "key1"
	req.EXPECT().RoutingKey().Return(routingKey)

	defaultBackend.EXPECT().Handle(conn, req).Return(nil)

	dispatcher.Dispatch(conn, wasabi.MsgTypeText, data)
}

func TestRouterDispatcher_DispatchByRoutingKey(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)
	req := mocks.NewMockRequest(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return req
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().Context().Return(context.Background())

	data := []byte("test data")

	// Test case 1: Request with existing routing key
	routingKey := "key1"
	req.EXPECT().RoutingKey().Return(routingKey)

	mockBackend := mocks.NewMockBackend(t)
	mockBackend.EXPECT().Handle(conn, req).Return(nil)
	dispatcher.backendMap[routingKey] = mockBackend

	dispatcher.Dispatch(conn, wasabi.MsgTypeText, data)
}

func TestRouterDispatcher_DispatchWrongRequest(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return nil
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().Context().Return(context.Background())

	data := []byte("test data")

	dispatcher.Dispatch(conn, wasabi.MsgTypeText, data)
}

func TestRouterDispatcher_DispatchErrorHandlingRequest(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)
	req := mocks.NewMockRequest(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return req
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	conn := mocks.NewMockConnection(t)
	conn.EXPECT().Context().Return(context.Background())

	data := []byte("test data")

	routingKey := "key1"
	req.EXPECT().RoutingKey().Return(routingKey)

	mockBackend := mocks.NewMockBackend(t)
	mockBackend.EXPECT().Handle(conn, req).Return(fmt.Errorf("test error"))
	dispatcher.backendMap[routingKey] = mockBackend

	dispatcher.Dispatch(conn, wasabi.MsgTypeText, data)
}
func TestRouterDispatcher_Use(t *testing.T) {
	defaultBackend := mocks.NewMockBackend(t)
	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return mocks.NewMockRequest(t)
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	middleware := RequestMiddlewere(func(next wasabi.RequestHandler) wasabi.RequestHandler { return next })
	dispatcher.Use(middleware)

	if len(dispatcher.middlewares) != 1 {
		t.Errorf("Expected middlewares length to be 1, but got %d", len(dispatcher.middlewares))
	}
}
func TestRouterDispatcher_UseMiddleware(t *testing.T) {
	testError := fmt.Errorf("test error")

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	defaultBackend := mocks.NewMockBackend(t)
	defaultBackend.EXPECT().Handle(mockConn, mockReq).Return(testError)

	parser := func(_ wasabi.Connection, _ context.Context, _ wasabi.MessageType, _ []byte) wasabi.Request {
		return mocks.NewMockRequest(t)
	}
	dispatcher := NewRouterDispatcher(defaultBackend, parser)

	middleware1 := RequestMiddlewere(func(next wasabi.RequestHandler) wasabi.RequestHandler { return next })
	middleware2 := RequestMiddlewere(func(next wasabi.RequestHandler) wasabi.RequestHandler { return next })

	dispatcher.Use(middleware1)
	dispatcher.Use(middleware2)

	endpointWithMiddleware := dispatcher.useMiddleware(defaultBackend)

	err := endpointWithMiddleware.Handle(mockConn, mockReq)

	if err != testError {
		t.Errorf("Expected error %v, but got %v", testError, err)
	}
}

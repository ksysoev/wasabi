package dispatch

import (
	"context"
	"fmt"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewPipeDispatcher(t *testing.T) {
	backend := mocks.NewMockBackend(t)
	dispatcher := NewPipeDispatcher(backend)

	if dispatcher.backend != backend {
		t.Errorf("Expected backend to be %v, but got %v", backend, dispatcher.backend)
	}

	if len(dispatcher.middlewares) != 0 {
		t.Errorf("Expected no middlewares, but got %d", len(dispatcher.middlewares))
	}
}

func TestPipeDispatcher_Dispatch(t *testing.T) {
	backend := mocks.NewMockBackend(t)
	dispatcher := NewPipeDispatcher(backend)

	conn := mocks.NewMockConnection(t)
	data := []byte("test data")
	testError := fmt.Errorf("test error")

	conn.On("Context").Return(context.Background())
	backend.EXPECT().Handle(conn, NewRawRequest(conn.Context(), wasabi.MsgTypeText, data)).Return(testError)

	dispatcher.Dispatch(conn, wasabi.MsgTypeText, data)
}

func TestPipeDispatcher_Use(t *testing.T) {
	backend := mocks.NewMockBackend(t)
	dispatcher := NewPipeDispatcher(backend)

	middleware := RequestMiddlewere(func(next wasabi.RequestHandler) wasabi.RequestHandler { return next })
	dispatcher.Use(middleware)

	if len(dispatcher.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, but got %d", len(dispatcher.middlewares))
	}
}

func TestPipeDispatcher_useMiddleware(t *testing.T) {
	backend := mocks.NewMockBackend(t)
	dispatcher := NewPipeDispatcher(backend)

	middleware := RequestMiddlewere(func(next wasabi.RequestHandler) wasabi.RequestHandler { return next })
	dispatcher.Use(middleware)

	handler := mocks.NewMockRequestHandler(t)
	result := dispatcher.useMiddleware(handler)

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

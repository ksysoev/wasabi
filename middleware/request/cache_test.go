package request

import (
	"context"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewCacheMiddleware(t *testing.T) {
	// Create a mock request handler
	testResp := []byte("testData")
	callCounter := 0
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		callCounter++
		return conn.Send(wasabi.MsgTypeText, testResp)
	})

	// Define the request cache function
	requestCache := func(r wasabi.Request) (cacheKey string, ttl time.Duration) {
		return "cacheKey", time.Second
	}

	// Create the cache middleware
	middleware, closer := NewCacheMiddleware(requestCache)
	defer closer()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(context.Background())

	// Test with cache hit
	mockConn.EXPECT().Send(wasabi.MsgTypeText, testResp).Return(nil).Times(2)
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCounter)

	err = middleware(mockHandler).Handle(mockConn, mockReq)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCounter)
}

func TestNewCacheMiddleware_NoCache(t *testing.T) {
	testResp := []byte("testData")
	callCounter := 0
	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		callCounter++
		return conn.Send(wasabi.MsgTypeText, testResp)
	})

	// Define the request cache function
	requestCache := func(r wasabi.Request) (cacheKey string, ttl time.Duration) {
		return "", 0
	}

	// Create the cache middleware
	middleware, closer := NewCacheMiddleware(requestCache)
	defer closer()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	// Test without caching
	mockConn.EXPECT().Send(wasabi.MsgTypeText, testResp).Return(nil).Times(2)
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCounter)

	err = middleware(mockHandler).Handle(mockConn, mockReq)
	assert.NoError(t, err)
	assert.Equal(t, 2, callCounter)
}

func TestNewCacheMiddleware_Error(t *testing.T) {
	// Create a mock request handler
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		return assert.AnError
	})

	// Define the request cache function
	requestCache := func(r wasabi.Request) (cacheKey string, ttl time.Duration) {
		return "cacheKey", time.Second
	}

	// Create the cache middleware
	middleware, closer := NewCacheMiddleware(requestCache)
	defer closer()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	// Test with error in handler
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	assert.EqualError(t, err, assert.AnError.Error())
}

func TestNewCacheMiddleware_ContextCancelled(t *testing.T) {
	// Create a mock request handler
	testResp := []byte("testData")
	callCounter := 0
	mockHandler := dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
		callCounter++
		return conn.Send(wasabi.MsgTypeText, testResp)
	})

	// Define the request cache function
	requestCache := func(r wasabi.Request) (cacheKey string, ttl time.Duration) {
		return "cacheKey", time.Second
	}

	// Create the cache middleware
	middleware, closer := NewCacheMiddleware(requestCache)
	defer closer()

	// Create a mock connection and request
	mockConn := mocks.NewMockConnection(t)
	ctx, cancel := context.WithCancel(context.Background())
	mockReq := mocks.NewMockRequest(t)
	mockReq.EXPECT().Context().Return(ctx)

	cancel()

	// Test with cancelled context
	err := middleware(mockHandler).Handle(mockConn, mockReq)
	assert.EqualError(t, err, context.Canceled.Error())
}

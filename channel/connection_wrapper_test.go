package channel

import (
	"context"
	"testing"

	"github.com/coder/websocket"
	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)

	wrapper := NewConnectionWrapper(mockConnection)

	assert.NotNil(t, wrapper)
	assert.Equal(t, mockConnection, wrapper.connection)
}

func TestConnectionWrapper_ID(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedID := "testID"
	mockConnection.On("ID").Return(expectedID)

	actualID := wrapper.ID()

	assert.Equal(t, expectedID, actualID)
}

func TestConnectionWrapper_Context(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedContext := context.TODO()
	mockConnection.On("Context").Return(expectedContext)

	actualContext := wrapper.Context()

	assert.Equal(t, expectedContext, actualContext)
}

func TestConnectionWrapper_Send_WithOnSendWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedMsgType := wasabi.MessageType(1)
	expectedMsg := []byte("test message")

	mockOnSendWrapper := func(conn wasabi.Connection, msgType wasabi.MessageType, msg []byte) error {
		assert.Equal(t, mockConnection, conn)
		assert.Equal(t, expectedMsgType, msgType)
		assert.Equal(t, expectedMsg, msg)

		return nil
	}

	wrapper.onSendWrapper = mockOnSendWrapper

	err := wrapper.Send(expectedMsgType, expectedMsg)

	assert.NoError(t, err)
}

func TestConnectionWrapper_Send_WithoutOnSendWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedMsgType := wasabi.MessageType(1)
	expectedMsg := []byte("test message")

	mockConnection.On("Send", expectedMsgType, expectedMsg).Return(nil)

	err := wrapper.Send(expectedMsgType, expectedMsg)

	assert.NoError(t, err)
	mockConnection.AssertExpectations(t)
}

func TestConnectionWrapper_Close_WithOnCloseWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedStatus := websocket.StatusCode(1000)
	expectedReason := "test reason"
	expectedClosingCtx := context.TODO()

	mockOnCloseWrapper := func(conn wasabi.Connection, status websocket.StatusCode, reason string, closingCtx ...context.Context) error {
		assert.Equal(t, mockConnection, conn)
		assert.Equal(t, expectedStatus, status)
		assert.Equal(t, expectedReason, reason)
		assert.Equal(t, expectedClosingCtx, closingCtx[0])

		return nil
	}

	wrapper.onCloseWrapper = mockOnCloseWrapper

	err := wrapper.Close(expectedStatus, expectedReason, expectedClosingCtx)

	assert.NoError(t, err)
}

func TestConnectionWrapper_Close_WithoutOnCloseWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	expectedStatus := websocket.StatusCode(1000)
	expectedReason := "test reason"
	expectedClosingCtx := context.TODO()

	mockConnection.On("Close", expectedStatus, expectedReason, expectedClosingCtx).Return(nil)

	err := wrapper.Close(expectedStatus, expectedReason, expectedClosingCtx)

	assert.NoError(t, err)
	mockConnection.AssertExpectations(t)
}

func TestConnectionWrapper_WithSendWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	if wrapper.onSendWrapper != nil {
		t.Error("Expected onSendWrapper to be nil")
	}

	cb := func(_ wasabi.Connection, _ wasabi.MessageType, _ []byte) error {
		return nil
	}

	wrapper = NewConnectionWrapper(mockConnection, WithSendWrapper(cb))

	if wrapper.onSendWrapper == nil {
		t.Error("Expected onSendWrapper to be set")
	}
}

func TestConnectionWrapper_WithCloseWrapper(t *testing.T) {
	mockConnection := mocks.NewMockConnection(t)
	wrapper := NewConnectionWrapper(mockConnection)

	if wrapper.onCloseWrapper != nil {
		t.Error("Expected onCloseWrapper to be nil")
	}

	cb := func(_ wasabi.Connection, _ websocket.StatusCode, _ string, _ ...context.Context) error {
		return nil
	}

	wrapper = NewConnectionWrapper(mockConnection, WithCloseWrapper(cb))

	if wrapper.onCloseWrapper == nil {
		t.Error("Expected onCloseWrapper to be set")
	}
}

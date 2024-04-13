package dispatch

import (
	"bytes"
	"context"
	"testing"

	"github.com/ksysoev/wasabi"
)

func TestRawRequest_Data(t *testing.T) {
	data := []byte("test data")
	req := NewRawRequest(context.Background(), wasabi.MsgTypeText, data)

	if !bytes.Equal(req.Data(), data) {
		t.Errorf("Expected data to be '%s', but got '%s'", data, req.Data())
	}
}

func TestRawRequest_RoutingKey(t *testing.T) {
	req := NewRawRequest(context.Background(), wasabi.MsgTypeText, []byte{})

	if req.RoutingKey() != "text" {
		t.Errorf("Expected routing key to be empty, but got %v", req.RoutingKey())
	}
}

func TestRawRequest_Context(t *testing.T) {
	ctx := context.Background()
	req := NewRawRequest(ctx, wasabi.MsgTypeText, []byte{})

	if req.Context() != ctx {
		t.Errorf("Expected context to be %v, but got %v", ctx, req.Context())
	}
}

func TestRawRequest_WithContext(t *testing.T) {
	ctx := context.Background()
	req := NewRawRequest(context.Background(), wasabi.MsgTypeText, []byte{})

	newReq := req.WithContext(ctx)

	if newReq.Context() != ctx {
		t.Errorf("Expected context to be %v, but got %v", ctx, newReq.Context())
	}

	if newReq != req {
		t.Error("Expected WithContext to return the same request instance")
	}
}

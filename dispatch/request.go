package dispatch

import (
	"context"

	"github.com/ksysoev/wasabi"
)

type RawRequest struct {
	ctx     context.Context
	data    []byte
	msgType wasabi.MessageType
}

func NewRawRequest(ctx context.Context, msgType wasabi.MessageType, data []byte) *RawRequest {
	if ctx == nil {
		panic("nil context")
	}

	return &RawRequest{ctx: ctx, data: data, msgType: msgType}
}

func (r *RawRequest) Data() []byte {
	return r.data
}

func (r *RawRequest) RoutingKey() string {
	switch r.msgType {
	case wasabi.MsgTypeText:
		return "text"
	case wasabi.MsgTypeBinary:
		return "binary"
	default:
		panic("unknown message type " + r.msgType.String())
	}
}

func (r *RawRequest) Context() context.Context {
	return r.ctx
}

func (r *RawRequest) WithContext(ctx context.Context) wasabi.Request {
	if ctx == nil {
		panic("nil context")
	}

	//TODO: Shall we copy the request? it feels that it will be a bit slow
	// but in http.Request they do it https://cs.opensource.google/go/go/+/master:src/net/http/request.go;l=362
	// for now we will just set the context and return original request
	// but I should think about it
	r.ctx = ctx

	return r
}

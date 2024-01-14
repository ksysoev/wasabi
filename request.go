package wasabi

import (
	"context"
)

type Request interface {
	Data() []byte
	RoutingKey() string
	Context() context.Context
	WithContext(ctx context.Context) Request
}

type RawRequest struct {
	ctx  context.Context
	data []byte
}

func NewRawRequest(ctx context.Context, data []byte) *RawRequest {
	if ctx == nil {
		panic("nil context")
	}

	return &RawRequest{ctx: ctx, data: data}
}

func (r *RawRequest) Data() []byte {
	return r.data
}

func (r *RawRequest) RoutingKey() string {
	return ""
}

func (r *RawRequest) Context() context.Context {
	return r.ctx
}

func (r *RawRequest) WithContext(ctx context.Context) Request {
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

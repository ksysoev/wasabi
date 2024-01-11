package wasabi

import (
	"context"
	"encoding/json"
	"fmt"
)

type Request interface {
	String() (string, error)
	RoutingKey() string
	Context() context.Context
	WithContext(ctx context.Context) Request
}

type RequestParser interface {
	Parse(data []byte) (Request, error)
}

type JSONRPCRequest struct {
	orginReq     *RPCRequest
	ctx          context.Context
	Data         string
	ID           string
	ConnectionID string
}

type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	// TODO make work with string and numbers
	ID string `json:"id"`
}
type JSONRPCRequestParser struct{}

func (j *JSONRPCRequestParser) Parse(data []byte) (Request, error) {
	msg := string(data)

	originReq, err := parseRequest(msg)

	if err != nil {
		return nil, err
	}

	return &JSONRPCRequest{
		ID:       originReq.ID,
		Data:     msg,
		orginReq: originReq,
	}, nil
}

func parseRequest(data string) (*RPCRequest, error) {
	var req RPCRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, err
	}

	if req.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid JSON-RPC version: %s", req.JSONRPC)
	}

	return &req, nil
}

func (r *JSONRPCRequest) String() (string, error) {
	return r.Data, nil
}

func (r *JSONRPCRequest) RoutingKey() string {
	return r.orginReq.Method
}

func (r *JSONRPCRequest) Context() context.Context {
	return r.ctx
}

func (r *JSONRPCRequest) WithContext(ctx context.Context) Request {
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

package wasabi

import (
	"encoding/json"
	"fmt"
)

type Request interface {
	String() (string, error)
	Action() string
	Stash() Stasher
}

type RequestParser interface {
	Parse(data []byte) (Request, error)
}

type JSONRPCRequest struct {
	Data         string
	orginReq     *RPCRequest
	ID           string
	ConnectionID string
	stash        Stasher
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
		stash:    NewStashStore(),
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

func (r *JSONRPCRequest) Action() string {
	return r.orginReq.Method
}

func (r *JSONRPCRequest) Stash() Stasher {
	return r.stash
}

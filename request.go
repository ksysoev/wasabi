package wasabi

import (
	"encoding/json"
	"fmt"
)

type JSONRPCRequest struct {
	Data         string
	orginReq     *RPCRequest
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

func NewRequest(data string) (*JSONRPCRequest, error) {
	originReq, err := parseRequest(data)

	if err != nil {
		return nil, err
	}

	return &JSONRPCRequest{ID: originReq.ID, Data: data, orginReq: originReq}, nil
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

package wasabi

import "encoding/json"

type JSONRPCResponse struct {
	id      string
	data    string
	isError bool
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      string      `json:"id"`
}

type RPCError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewResponse(id, data string) *JSONRPCResponse {
	return &JSONRPCResponse{id: id, data: data, isError: false}
}

func ResponseFromError(err error) *JSONRPCResponse {
	return &JSONRPCResponse{data: err.Error(), isError: true}
}

func (r *JSONRPCResponse) String() (string, error) {
	resp := RPCResponse{
		JSONRPC: "2.0",
		ID:      r.id,
	}

	if r.isError {
		resp.Error = &RPCError{
			Code:    1,
			Message: r.data,
		}
	} else {
		resp.Result = r.data
	}

	data, err := json.Marshal(resp)

	return string(data), err
}

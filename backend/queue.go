package backend

import (
	"github.com/google/uuid"
	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type response struct {
	msgType websocket.MessageType
	data    []byte
}

type OnRequestCallback func(conn wasabi.Connection, req wasabi.Request, id string) error

type QueueBackend struct {
	requests  map[string]chan response
	onRequest OnRequestCallback
}

func NewQueueBackend(onRequest OnRequestCallback) *QueueBackend {
	return &QueueBackend{
		requests:  make(map[string]chan response),
		onRequest: onRequest,
	}
}

func (b *QueueBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	id := uuid.New().String()
	respChan := make(chan response)

	b.requests[id] = respChan

	err := b.onRequest(conn, r, id)

	if err != nil {
		return err
	}

	select {
	case resp := <-respChan:
		return conn.Send(resp.msgType, resp.data)
	case <-r.Context().Done():
		return nil
	}
}

func (b *QueueBackend) OnResponse(id string, msgType websocket.MessageType, data []byte) {
	b.requests[id] <- response{msgType, data}
	close(b.requests[id])
}

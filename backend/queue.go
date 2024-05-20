package backend

import (
	"strconv"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type response struct {
	data    []byte
	msgType websocket.MessageType
}

type OnRequestCallback func(conn wasabi.Connection, req wasabi.Request, id string) error

type QueueBackend struct {
	requests  map[string]chan response
	onRequest OnRequestCallback
	lock      *sync.Mutex
	lastReqID int
}

func NewQueueBackend(onRequest OnRequestCallback) *QueueBackend {
	return &QueueBackend{
		requests:  make(map[string]chan response),
		onRequest: onRequest,
		lock:      &sync.Mutex{},
		lastReqID: 1,
	}
}

func (b *QueueBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	respChan := make(chan response)

	b.lock.Lock()
	b.lastReqID++
	id := strconv.Itoa(b.lastReqID)
	b.requests[id] = respChan
	b.lock.Unlock()

	defer func() {
		b.lock.Lock()
		delete(b.requests, id)
		b.lock.Unlock()
	}()

	err := b.onRequest(conn, r, id)

	if err != nil {
		return err
	}

	select {
	case resp := <-respChan:
		return conn.Send(resp.msgType, resp.data)
	case <-r.Context().Done():
		return r.Context().Err()
	}
}

func (b *QueueBackend) OnResponse(id string, msgType websocket.MessageType, data []byte) {
	b.lock.Lock()
	respChan, ok := b.requests[id]
	b.lock.Unlock()

	if !ok {
		return
	}

	defer close(respChan)

	select {
	case respChan <- response{data, msgType}:
	default:
	}
}

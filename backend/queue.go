package backend

import (
	"context"
	"strconv"
	"sync"

	"github.com/ksysoev/wasabi"
	"nhooyr.io/websocket"
)

type response struct {
	data    []byte
	msgType websocket.MessageType
}

type request struct {
	respChan chan response
	ctx      context.Context
}

// OnRequestCallback is a function type that represents a callback function
// for handling requests in the queue.
// It takes three parameters:
//   - conn: a `wasabi.Connection` object representing the connection.
//   - req: a `wasabi.Request` object representing the request.
//   - id: a string representing the ID of the request.
//
// It returns an error if there was an issue handling the request.
type OnRequestCallback func(conn wasabi.Connection, req wasabi.Request, id string) error

// QueueBackend represents a backend for handling requests in a queue.
type QueueBackend struct {
	requests  map[string]request
	onRequest OnRequestCallback
	lock      *sync.Mutex
	lastReqID int
}

// NewQueueBackend creates a new instance of QueueBackend.
// It takes an onRequest callback function as a parameter and returns a pointer to QueueBackend.
// The onRequest callback function is called when a new request is received  and ready to be passed to queue.
func NewQueueBackend(onRequest OnRequestCallback) *QueueBackend {
	return &QueueBackend{
		requests:  make(map[string]request),
		onRequest: onRequest,
		lock:      &sync.Mutex{},
		lastReqID: 1,
	}
}

// Handle handles the incoming request from the given connection.
// It processes the request, sends the response back to the connection,
// and returns any error that occurred during the handling process.
func (b *QueueBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	respChan := make(chan response)

	b.lock.Lock()
	b.lastReqID++
	id := strconv.Itoa(b.lastReqID)
	b.requests[id] = request{respChan, r.Context()}
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

// OnResponse handles the response received from the server for a specific request.
// It takes the ID of the request, the message type, and the response data as parameters.
// If there is a corresponding request channel for the given ID, it sends the response
// to the channel. If nobody is awaiting for response, it discards the response.
func (b *QueueBackend) OnResponse(id string, msgType websocket.MessageType, data []byte) {
	b.lock.Lock()
	request, ok := b.requests[id]
	b.lock.Unlock()

	if !ok {
		return
	}

	defer close(request.respChan)

	select {
	case request.respChan <- response{data, msgType}:
	case <-request.ctx.Done():
	}
}

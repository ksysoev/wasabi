package dispatch

import (
	"fmt"
	"log/slog"

	"github.com/ksysoev/wasabi"
)

type RouterDispatcher struct {
	defaultBackend wasabi.RequestHandler
	backendMap     map[string]wasabi.RequestHandler
	parser         RequestParser
	middlewares    []RequestMiddlewere
}

// NewRouterDispatcher creates a new instance of RouterDispatcher.
// It takes a defaultBackend and a request parser as parameters and returns a pointer to RouterDispatcher.
// The defaultBackend parameter is the default backend to be used when no specific backend is found.
// The parser parameter is used to parse incoming requests.
func NewRouterDispatcher(defaultBackend wasabi.RequestHandler, parser RequestParser) *RouterDispatcher {
	return &RouterDispatcher{
		defaultBackend: defaultBackend,
		backendMap:     make(map[string]wasabi.RequestHandler),
		parser:         parser,
	}
}

// AddBackend adds a backend to the RouterDispatcher for the specified routing keys.
// If a backend already exists for any of the routing keys, an error is returned.
func (d *RouterDispatcher) AddBackend(backend wasabi.RequestHandler, routingKeys []string) error {
	for _, key := range routingKeys {
		if _, ok := d.backendMap[key]; ok {
			return fmt.Errorf("backend for routing key %s already exists", key)
		}

		d.backendMap[key] = backend
	}

	return nil
}

// Dispatch handles the incoming connection and data by parsing the request,
// determining the appropriate backend, and handling the request using middleware.
// If an error occurs during handling, it is logged.
func (d *RouterDispatcher) Dispatch(conn wasabi.Connection, msgType wasabi.MessageType, data []byte) {
	req := d.parser(conn, conn.Context(), msgType, data)

	if req == nil {
		return
	}

	backend, ok := d.backendMap[req.RoutingKey()]
	if !ok {
		backend = d.defaultBackend
	}

	err := d.useMiddleware(backend).Handle(conn, req)
	if err != nil {
		slog.Error("Error handling request: " + err.Error())
	}
}

// Use adds a middleware to the router dispatcher.
// Middleware functions are executed in the order they are added.
func (d *RouterDispatcher) Use(middlewere RequestMiddlewere) {
	d.middlewares = append(d.middlewares, middlewere)
}

// useMiddleware applies the registered middlewares to the given endpoint.
// It iterates through the middlewares in reverse order and wraps the endpoint
// with each middleware function. The wrapped endpoint is then returned.
func (d *RouterDispatcher) useMiddleware(endpoint wasabi.RequestHandler) wasabi.RequestHandler {
	for i := len(d.middlewares) - 1; i >= 0; i-- {
		endpoint = d.middlewares[i](endpoint)
	}

	return endpoint
}

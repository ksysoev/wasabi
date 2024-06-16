# Wasabi: A Toolkit for Creating WebSocket API Gateway

[![WaSAbi](https://github.com/ksysoev/wasabi/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/wasabi/actions/workflows/main.yml)
[![CodeCov](https://codecov.io/gh/ksysoev/wasabi/graph/badge.svg?token=3KGTO1UINI)](https://codecov.io/gh/ksysoev/wasabi)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/wasabi)](https://goreportcard.com/report/github.com/ksysoev/wasabi)
[![Go Reference](https://pkg.go.dev/badge/github.com/ksysoev/wasabi.svg)](https://pkg.go.dev/github.com/ksysoev/wasabi)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

<p align="center">
    <img src="./logo.png" width="200px">
</p>

Wasabi is a Go package that provides a comprehensive toolkit for creating WebSocket API gateways. It provides a simple and intuitive API to build robust and scalable WebSocket applications.

**Note the package is still under active development, public interfaces are not stable and not production-ready yet**

## Installation

To install Wasabi, use the go get command:

```sh
go get github.com/ksysoev/wasabi
```

# Usage

Here's a basic example of how to use Wasabi to create a WebSocket API gateway:

```go
package main

import (
    "bytes"
    "context"
    "net/http"
    "os"

    "github.com/ksysoev/wasabi"
    "github.com/ksysoev/wasabi/backend"
    "github.com/ksysoev/wasabi/channel"
    "github.com/ksysoev/wasabi/dispatch"
    "github.com/ksysoev/wasabi/middleware/request"
    "github.com/ksysoev/wasabi/server"
)

const (
    Port = 8080
)

func main() {
    // We create a new backend with backend.NewBackend. 
    // This backend creates a new HTTP request for each incoming WebSocket message. 
    // The requests are sent to http://localhost:8081/.
    backend := backend.NewBackend(func(req wasabi.Request) (*http.Request, error) {
        httpReq, err := http.NewRequest("GET", "http://localhost:8081/", bytes.NewBuffer(req.Data()))
        if err != nil {
            return nil, err
        }

        return httpReq, nil
    })

    // We create an error handling middleware with request.NewErrorHandlingMiddleware. 
    // This middleware logs any errors that occur when handling a request and sends a response back to the client.
    ErrHandler := request.NewErrorHandlingMiddleware(func(conn wasabi.Connection, req wasabi.Request, err error) error {
        conn.Send([]byte("Failed to process request: " + err.Error()))
        return nil
    })

    // We create a new dispatcher with dispatch.NewPipeDispatcher. 
    // This dispatcher sends/routes each incoming WebSocket message to the backend.
    dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(conn.Context(), msgType, data)
	})
    
    dispatcher.Use(ErrHandler)
    dispatcher.Use(request.NewTrottlerMiddleware(10))

     // We create a new connection registry with channel.NewConnectionRegistry. 
     // This registry keeps track of all active connections 
     // and responsible for managing connection's settings.
     connRegistry := channel.NewConnectionRegistry()

    // We create a new server with wasabi.NewServer and add a channel to it with server.AddChannel. 
    // The server listens on port 8080 and the channel handles all requests to the / path.
    channel := channel.NewChannel("/", dispatcher, connRegistry)
    server := server.NewServer(Port)
    server.AddChannel(channel)

    // Finally, we start the server with server.Run. 
    // If the server fails to start, we log the error and exit the program.
    if err := server.Run(); err != nil {
        slog.Error("Fail to start app server", "error", err)
        os.Exit(1)
    }

    os.Exit(0)
}
```


## Core concepts

- **Server**: This is the main component that listens for incoming HTTP requests. It manages channels and dispatches requests to them.
- **Channel**: A channel represents an endpoint for WebSocket connections. It's responsible for handling all WebSocket connections and messages for a specific path.
- **Dispatcher**: This acts as a router for incoming WebSocket messages. It uses middleware to process messages and dispatches them to the appropriate backend.
- **Backend**: This is the handler for WebSocket messages. Once a message has been processed by the dispatcher and any middleware, it's sent to the backend for further processing.
- **Connection Registry**: This is a central registry for managing WebSocket connections. It keeps track of all active connections and their settings.
- **Connection**: This represents a single WebSocket connection. It's managed by the connection registry and used by the server and channel to send and receive messages.
- **Request**: This represents a single WebSocket message. It's created by the server when a new message is received and processed by the dispatcher and backend.
- **Middleware** is used to process messages before they reach the backend. There are two types of middleware:
  - **HTTP Middleware**: This processes incoming HTTP requests before they're upgraded to WebSocket connections.
  - **Request Middleware**: This processes incoming WebSocket messages before they're dispatched to the backend.


### Server

The Server is the main component of the library. It listens for incoming HTTP requests, manages channels, and dispatches requests to them. The server is responsible for starting the WebSocket service and managing its lifecycle.

When a new server is created with `server.NewServer`, it's initialized with a port number. This is the port that the server will listen on for incoming HTTP requests.

```golang
server := server.NewServer(":8080")
```

Channels are added to the server with `server.AddChannel`. Each channel represents a different WebSocket endpoint. The server will dispatch incoming WebSocket requests to the appropriate channel based on the request path.

```golang
channel := channel.NewChannel("/", dispatcher, connRegistry)
server.AddChannel(channel)
```

The server is started with `server.Run`. This method takes a context, which is used to control the server's lifecycle. If the server fails to start for any reason, `server.Run` will return an error.

```golang
if err := server.Run(); err != nil {
    slog.Error("Fail to start app server", "error", err)
    os.Exit(1)
}
```

In this example, if the server fails to start, the error is logged and the program is exited with a non-zero status code.

The server is a crucial part of the WebSocket service. It's responsible for managing the service's lifecycle and dispatching HTTP requests to the appropriate channels.


### Channel

### Dispatcher

## Contributing

Contributions to Wasabi are welcome! Please submit a pull request or create an issue to contribute.

For an easy start please read [the contributor's guidelines](./CONTRIBUTING.md).

# License

Wasabi is licensed under the MIT License. See the LICENSE file for more details.
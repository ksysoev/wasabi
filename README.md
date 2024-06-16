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
        conn.Send(wasabi.MsgTypeTex, []byte("Failed to process request: " + err.Error()))
        return nil
    })

    // We create a new dispatcher with dispatch.NewPipeDispatcher. 
    // This dispatcher sends/routes each incoming WebSocket message to the backend.
    dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
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
import "github.com/ksysoev/wasabi/server"

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

A Channel in the context of WebSocket connections serves as an endpoint. It is responsible for managing WebSocket connections and messages for a specific path.

Key responsibilities of the Channel abstraction include:

- Processing client requests to establish WebSocket connections.
- Storing configuration required for establishing WebSocket connections.
- Managing and executing middleware for HTTP requests.


When a new channel is created with `channel.NewChannel`, it is initialized with a path, a dispatcher, and a connection registry. The path is the URL path that the channel will handle. The dispatcher is used to process incoming WebSocket messages, and the connection registry is used to manage active WebSocket connections.

```golang
import "github.com/ksysoev/wasabi/channel"

chatChan := channel.NewChannel("/chat", dispatcher, connRegistry)
```

In this example, a new channel is created to handle WebSocket connections on the `/chat` path.

Channels are added to a server with `server.AddChannel`. The server will dispatch incoming WebSocket requests to the appropriate channel based on the request path.

```golang
server.AddChannel(chatChan)
```

In this example, the channel is added to the server. Any incoming WebSocket requests on the `/chat` path will be handled by this channel.

### Connection Registry

The Connection Registry is responsible for:

- Managing the lifecycle of connections.
- Defining WebSocket connections configurations.
- Managing hooks for establishing and closing client connections.

```golang 
import "github.com/ksysoev/wasabi/channel"

connRegistry := channel.NewConnectionRegistry()
```

In this example, a new connection registry is created.

### Connection

A Connection represents an active WebSocket connection. It provides methods for sending messages and closing the connection.

To send a message, use the `Send` method. This method takes a message type and a bytes slice as arguments.

```golang
err := conn.Send(wasabi.MsgTypeText, "Hello World!")
```

In this example, a text message "Hello World!" is being sent over the WebSocket connection.

To close a WebSocket connection, use the `Close` method. This method takes a status code and a string reason as arguments.

```golang
conn.Close(websocket.StatusGoingAway, "Server is restarting")
```

In this example, the WebSocket connection is being closed with a status code indicating that the server is going away and a reason "Server is restarting".

### Dispatcher

A Dispatcher acts as a router for incoming WebSocket messages. It uses middleware to process messages and dispatches them to the appropriate backend.

When a new dispatcher is created with `dispatcher.NewRouterDispatcher`, it's initialized with a default backend and request parser. The backend is the handler for WebSocket messages. Once a message has been processed by the dispatcher and any middleware, it's sent to the backend for further processing.

```golang
import "github.com/ksysoev/wasabi/dispatcher"

chatDipatcher := dispatcher.NewRouterDispatcher(
    myBackend, 
    func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
    },
)
```

In this example, a new dispatcher is created with a custom backend that is stored in `myBackend` variable. The second argument is the request parser that accepts WebSocket messages and returns Request.

The router dispatcher allows to routing of incoming WebSocket messages to multiple backends, to add additional backends to the created dispatcher you can use `channel.AddBackend` method:

```golang
chatDipatcher.AddBackend(myNotificationBackend, []string{"notifications", "subscriptions"})
```

In this example, we're adding a backend to the chatDispatcher. The backend is named myNotificationBackend and it's being associated with two routing keys: "notifications" and "subscriptions".

The dispatcher is responsible for processing WebSocket messages and dispatching them to the appropriate backend.

# Request

A Request represents a single WebSocket message. It encapsulates the data and metadata of a WebSocket message that is to be processed by the dispatcher and backend.

To allow integration with the dispatcher and backend abstractions, the request structure should implement the wasabi.Request interface. This interface ensures that the request has the necessary methods for handling and processing.

- `Context`: This method returns the context of the request. It can be used to carry request-scoped values, cancellation signals, and deadlines across API boundaries and between processes.
- `Data`: This method returns the data of the WebSocket message. This is the actual content of the message that needs to be processed.
- `RoutingKey`: This method returns the routing key that will be used for routing the message to the correct backend.
- `WithContext(ctx context.Context)`: This method is used to assign an adjusted context to the request. It's useful when you want to propagate a new derived context to the request.

Here's an example of a custom request structure that implements the wasabi.Request interface:

```golang
type MyRequest struct {
    ctx context.Context
    msgType wasabi.MessageType
    data []byte
    routingKey string
}

func (r *MyRequest) Context() context.Context {
    return r.ctx
}

func (r *MyRequest) Data() []byte {
    return r.data
}

func (r *MyRequest) RoutingKey() string {
    return r.routingKey
}

func (r *MyRequest) WithContext(ctx context.Context) wasabi.Request {
    r.ctx = ctx
    return r
}
```

In this example, `MyRequest` implements the `wasabi.Request` interface. It can now be used with the dispatcher and backend abstractions to process WebSocket messages.

### Backend 

A Backend is the handler for WebSocket messages. After a message has been processed by the dispatcher and any middleware, it's forwarded to the backend for further processing.

To integrate a backend with the dispatcher, it should implement the `wasabi.RequestHandler` interface. This interface ensures that the backend has the necessary method for handling requests.

Wasabi provides several predefined backends that can be used directly:

- **HTTP backend**: This backend is designed for integration with HTTP application servers. It allows WebSocket messages to be processed and responded to using standard HTTP request handling.
- **WebSocket backend**: This backend is designed for integration with WebSocket applications. It provides a direct interface for processing WebSocket messages.
- **Queue backend**: This backend is designed for integration with backend applications via queue systems like RabbitMQ, Redis, Kafka, etc. It allows WebSocket messages to be processed asynchronously and in a distributed manner.

Here's an example of creating an HTTP backend:

```golang
backend := backend.NewBackend(func(req wasabi.Request) (*http.Request, error) {
    httpReq, err := http.NewRequest("POST", "http://localhost:8081/", bytes.NewBuffer(req.Data()))
    if err != nil {
        return nil, err
    }

    return httpReq, nil
})
```

In this code example, we're creating an HTTP backend to integrate with our application service. The backend takes a WebSocket request, creates a new HTTP request with the same data, and returns the HTTP request for further processing.


## Contributing

Contributions to Wasabi are welcome! Please submit a pull request or create an issue to contribute.

For an easy start please read [the contributor's guidelines](./CONTRIBUTING.md).

# License

Wasabi is licensed under the MIT License. See the LICENSE file for more details.
# Wasabi: A Toolkit for Creating WebSocket API Gateway

[![WaSAbi](https://github.com/ksysoev/wasabi/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/wasabi/actions/workflows/main.yml)
[![CodeCov](https://codecov.io/gh/ksysoev/wasabi/graph/badge.svg?token=3KGTO1UINI)](https://codecov.io/gh/ksysoev/wasabi)
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
    if err := server.Run(context.Background()); err != nil {
        slog.Error("Fail to start app server", "error", err)
        os.Exit(1)
    }

    os.Exit(0)
}
```

## Contributing

Contributions to Wasabi are welcome! Please submit a pull request or create an issue to contribute.

# License

Wasabi is licensed under the MIT License. See the LICENSE file for more details.
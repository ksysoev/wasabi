package channel

import "net/http"

// Middlewere is interface for middlewares
type Middlewere func(http.Handler) http.Handler

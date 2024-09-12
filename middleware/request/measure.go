package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

// NewMeasurer returns a middleware function that measures the duration of each request and saves the metric using the provided `saveMetric` function.
// The `saveMetric` function takes three parameters: the request, the error (if any), and the duration of the request.
// The middleware function wraps the provided `next` request handler and measures the duration of the request by calculating the time elapsed between the start and end of the request.
// After the request is handled, the metric is saved using the `saveMetric` function.
// The middleware function returns the wrapped request handler.
func NewMeasurer(saveMetric func(req wasabi.Request, err error, duration time.Duration)) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			start := time.Now()
			err := next.Handle(conn, req)
			duration := time.Since(start)

			saveMetric(req, err, duration)

			return err
		})
	}
}

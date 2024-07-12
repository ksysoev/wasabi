package request

import (
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// NewSpanMiddleware returns a middleware function that creates a span for its requests.
// It takes the `spanName` and `attributes` as an argument. The attributes are essentially key value tuples used to set
// the span's attributes during creation.
// This middleware must be used in conjunction after the `WithTracer` server option
func NewSpanMiddleware(spanName string, tracer trace.Tracer, attributes ...attribute.KeyValue) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			if tracer == nil {
				panic("NewSpanMiddleware called without initializing Tracer! Have you specified `WithTracer` server option?")
			}

			ctx, span := tracer.Start(req.Context(), spanName, trace.WithAttributes(attributes...))
			defer span.End()

			span.AddEvent("span created", trace.WithAttributes(attribute.Int64("createdAt", time.Now().UnixMilli())))

			return next.Handle(conn, req.WithContext(ctx))
		})
	}
}

package request

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// NewSpanMiddleware returns a middleware function that creates a span for its requests.
// It takes the `spanName` and `attributes` as an argument. The attributes are essentially key value tuples used to set
// the span's attributes during creation.
// This middleware must be used in conjunction after the `NewTracedMiddleware`
func NewSpanMiddleware(spanName string, tracer trace.Tracer, attributes ...attribute.KeyValue) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tracer == nil {
				panic("NewSpanMiddleware called without initializing Tracer! Have you specified `WithTracer` server option?")
			}

			ctx, span := tracer.Start(r.Context(), spanName, trace.WithAttributes(attributes...))
			defer span.End()

			span.AddEvent("span created", trace.WithAttributes(attribute.Int64("createdAt", time.Now().UnixMilli())))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

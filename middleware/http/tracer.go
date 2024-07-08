package http

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

// NewTracedMiddleware returns a middleware function that initializes an opentelemtry trace for its requests.
// It takes an `exporter` function as an argument. This function is used to set the otel exporter of your choice for tracing.
// Please refer to https://opentelemetry.io/docs/languages/go/exporters/ to know choices.
// Adding a traced middleware is a must if you want to create a span for your requests.
func NewTracedMiddleware(exporter func(ctx context.Context) (sdktrace.SpanExporter, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// setup opentelemtry exporter and provider
			exp, err := exporter(r.Context())
			if err != nil {
				log.Fatalf("failed to initialize exporter: %v", err)
			}

			tp := newTraceProvider(exp)

			defer func() { _ = tp.Shutdown(r.Context()) }()

			otel.SetTracerProvider(tp)

			// init tracer
			Tracer = tp.Tracer("Wasabi")

			// finally move to next middleware
			next.ServeHTTP(w, r)
		})
	}
}

// This merge logic ensures that semconv works as expected and that the version from the exportor matches the provider
// please see - https://github.com/open-telemetry/opentelemetry-go/issues/4476
func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("Wasabi"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}

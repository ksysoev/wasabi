package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/backend"
	"github.com/ksysoev/wasabi/channel"
	"github.com/ksysoev/wasabi/dispatch"
	whttp "github.com/ksysoev/wasabi/middleware/http"
	"github.com/ksysoev/wasabi/middleware/request"
	"github.com/ksysoev/wasabi/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer
var traceProvider *sdktrace.TracerProvider

const (
	Addr = ":8080"
)

func main() {
	slog.LogAttrs(context.Background(), slog.LevelDebug, "")

	backend := backend.NewBackend(func(req wasabi.Request) (*http.Request, error) {
		httpReq, err := http.NewRequest("POST", "http://localhost:8081/", bytes.NewBuffer(req.Data()))
		if err != nil {
			slog.Error("GOT ERROR AS - %s", err)
			return nil, err
		}

		return httpReq, nil
	})

	buildTrace(func(_ context.Context) (sdktrace.SpanExporter, error) { return stdouttrace.New() })

	protectedMiddleware := whttp.NewProtectedMiddleware(func(token string) error {
		if token != "secret" {
			return errors.New("token mismatch error")
		}
		return nil
	})
	spanMiddleware := request.NewSpanMiddleware("test-span", tracer)

	dispatcher := dispatch.NewRouterDispatcher(backend, func(conn wasabi.Connection, ctx context.Context, msgType wasabi.MessageType, data []byte) wasabi.Request {
		return dispatch.NewRawRequest(ctx, msgType, data)
	})
	dispatcher.Use(spanMiddleware)

	channel := channel.NewChannel("/", dispatcher, channel.NewConnectionRegistry(), channel.WithOriginPatterns("*"))
	channel.Use(protectedMiddleware)

	server := server.NewServer(Addr, server.WithBaseContext(context.Background()))
	server.AddChannel(channel)

	if err := server.Run(); err != nil {
		slog.Error("Fail to start app server", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func buildTrace(exporter func(ctx context.Context) (sdktrace.SpanExporter, error)) {
	exp, err := exporter(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize exporter: %v", err))
	}

	traceProvider = newTraceProvider(exp)
	otel.SetTracerProvider(traceProvider)
	tracer = traceProvider.Tracer("Test-Telemetry")
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

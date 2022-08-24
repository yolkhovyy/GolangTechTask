package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

type TraceProvider struct {
	traceProvider *trace.TracerProvider
}

func StartTrace() (shutdown func() error, err error) {
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/instrumentation/google.golang.org/grpc/otelgrpc/example/v0.33.0/instrumentation/google.golang.org/grpc/otelgrpc/example/config/config.go
	exporter, err := stdouttrace.New()
	if err != nil {
		return nil, err
	}
	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	shutdown = func() error {
		return traceProvider.Shutdown(context.Background())
	}
	return shutdown, nil
}

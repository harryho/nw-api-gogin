package telemetry

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
)

// Config captures the OpenTelemetry settings used by the API.
type Config struct {
	Endpoint       string
	Headers        map[string]string
	Insecure       bool
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Setup configures the global trace provider using the supplied configuration.
// When tracing is disabled (no endpoint), a no-op shutdown function is returned.
func Setup(ctx context.Context, cfg Config, log *zap.Logger) (func(context.Context) error, error) {
	if log == nil {
		log = zap.NewNop()
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		log.Info("otel exporter endpoint not configured; tracing disabled")
		return func(context.Context) error { return nil }, nil
	}

	clientOptions := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
	if cfg.Insecure {
		clientOptions = append(clientOptions, otlptracegrpc.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		clientOptions = append(clientOptions, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(clientOptions...))
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

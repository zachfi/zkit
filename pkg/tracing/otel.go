package tracing

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pkg/errors"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InstallOpenTelemetryTracer(config *Config, logger *slog.Logger, appName, version string) (func(), error) {
	if config.OtelEndpoint == "" {
		return func() {}, nil
	}

	logger.Info("initializing OpenTelemetry tracer", "endpoint", config.OtelEndpoint)

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(appName),
			semconv.ServiceVersionKey.String(version),
		),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize trace resource")
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	instrumentationOpts, err := config.GRPCClientConfig.DialOption(nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get dial options: %w", err)
	}

	opts = append(opts, instrumentationOpts...)

	conn, err := grpc.NewClient(config.OtelEndpoint, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial otel grpc")
	}

	options := []otlptracegrpc.Option{otlptracegrpc.WithGRPCConn(conn)}
	if config.OrgID != "" {
		options = append(options,
			otlptracegrpc.WithHeaders(map[string]string{"X-Scope-OrgID": config.OrgID}))
	}

	traceExporter, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to creat trace exporter")
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error("OpenTelemetry trace provider failed to shutdown", "err", err)
			os.Exit(1)
		}
	}

	return shutdown, nil
}

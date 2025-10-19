package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type Config struct {
	Server struct {
		Host string
		Port int
	}
	Metrics struct {
		Port     int
		Exporter string
	}
	Logging struct {
		Level  string
		Format string
	}
}

var cfg Config

func initConfig() error {
	// Server flags
	flag.StringVar(&cfg.Server.Host, "host", "localhost", "Server host name or IP address")
	flag.IntVar(&cfg.Server.Port, "port", 29070, "Server port")

	// Metrics flags
	flag.IntVar(&cfg.Metrics.Port, "metrics-port", 8870, "Metrics server port")
	flag.StringVar(&cfg.Metrics.Exporter, "exporter", "prometheus", "Metrics exporter type (prometheus, otlphttp, or otlpgrpc)")

	// Logging flags
	flag.StringVar(&cfg.Logging.Level, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.Logging.Format, "log-format", "text", "Log format (text or json)")

	flag.Parse()
	return nil
}

func initLogger(level string, format string) error {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s (must be 'debug', 'info', 'warn', or 'error')", level)
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel == slog.LevelDebug,
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return fmt.Errorf("invalid log format: %s (must be 'text' or 'json')", format)
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

func initOtel(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	instanceID := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	res, err := resource.New(ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName("jka-server"),
			semconv.ServiceInstanceID(instanceID),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var reader sdkmetric.Reader

	switch cfg.Metrics.Exporter {
	case "otlphttp":
		exporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP HTTP exporter: %w", err)
		}
		reader = sdkmetric.NewPeriodicReader(exporter)

	case "otlpgrpc":
		exporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP gRPC exporter: %w", err)
		}
		reader = sdkmetric.NewPeriodicReader(exporter)

	case "prometheus":
		exporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
		}
		reader = exporter

	default:
		return nil, fmt.Errorf("invalid exporter type: %s (must be 'prometheus', 'otlphttp', or 'otlpgrpc')", cfg.Metrics.Exporter)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	return provider, nil
}

func initMetrics(provider *sdkmetric.MeterProvider, q3connector *Q3Connector) error {
	meter := provider.Meter("jka-exporter")

	// Create all metrics
	var err error
	var currentClients, maxClients, playerPing metric.Int64ObservableUpDownCounter

	if currentClients, err = meter.Int64ObservableUpDownCounter(
		"jka.clients.connected",
		metric.WithDescription("Current number of clients connected"),
	); err != nil {
		return fmt.Errorf("failed to create jka.clients.connected metric: %w", err)
	}

	if maxClients, err = meter.Int64ObservableUpDownCounter(
		"jka.clients.max",
		metric.WithDescription("Maximum number of clients allowed"),
	); err != nil {
		return fmt.Errorf("failed to create jka.clients.max metric: %w", err)
	}

	if playerPing, err = meter.Int64ObservableUpDownCounter(
		"jka.clients.ping",
		metric.WithDescription("Player ping in milliseconds"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("failed to create jka.clients.ping metric: %w", err)
	}

	// Register callback for all metrics
	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		status, err := q3connector.GetStatus()
		if err != nil {
			return fmt.Errorf("failed to get server status: %w", err)
		}

		slog.Debug("server status retrieved", "status", fmt.Sprintf("%+v", status.Values))

		o.ObserveInt64(currentClients, int64(len(status.Players)))

		if maxClientsStr, ok := status.Values["sv_maxclients"]; ok {
			if maxClientsVal, err := strconv.Atoi(maxClientsStr); err == nil {
				o.ObserveInt64(maxClients, int64(maxClientsVal))
			}
		}

		for _, player := range status.Players {
			o.ObserveInt64(playerPing, int64(player.Ping),
				metric.WithAttributes(attribute.String("player", player.SanitizedName)))
		}
		return nil
	}, currentClients, maxClients, playerPing)

	if err != nil {
		return fmt.Errorf("failed to register metrics callback: %w", err)
	}

	return nil
}

func run() error {
	if err := initConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	if err := initLogger(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	ctx := context.Background()
	provider, err := initOtel(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
	}
	defer provider.Shutdown(ctx)

	connector := NewQ3Connector(cfg.Server.Host, cfg.Server.Port)
	if err := connector.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer connector.Close()

	if err := initMetrics(provider, connector); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	if cfg.Metrics.Exporter == "prometheus" {
		http.Handle("/metrics", promhttp.Handler())
	}

	slog.Info("starting metrics server",
		"port", cfg.Metrics.Port,
		"exporter", cfg.Metrics.Exporter,
	)
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Metrics.Port), nil)
}

func main() {
	if err := run(); err != nil {
		slog.Error("exporter failed", "error", err)
		os.Exit(1)
	}
}

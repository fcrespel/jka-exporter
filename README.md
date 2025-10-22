# JKA Exporter

A metrics exporter for Star Wars Jedi Knight: Jedi Academy dedicated servers, supporting Prometheus and OTLP formats.

## Usage

The following commands can be used to run the exporter as a [Docker](https://docs.docker.com/engine/) container.

```bash
# Show syntax help
docker run --rm ghcr.io/fcrespel/jka-exporter:master -help

# Start in the background with Prometheus exporter (default)
docker run -d --name jka-exporter -p 8870:8870 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070

# Start in the background with OTLP HTTP exporter
docker run -d --name jka-exporter -p 8870:8870 -e OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp-receiver:4318 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070 -exporter otlphttp

# Start in the background with OTLP gRPC exporter
docker run -d --name jka-exporter -p 8870:8870 -e OTEL_EXPORTER_OTLP_ENDPOINT=otlp-receiver:4317 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070 -exporter otlpgrpc

# Stop container
docker stop jka-exporter

# Delete container
docker rm jka-exporter
```

### Options

The following command line arguments are supported:

```
-host string
      Server host name or IP address (default "localhost")
-port int
      Server port (default 29070)
-rcon-password string
      Server Rcon password
-enable-rpmetrics
      Enable RPMod rpmetrics Rcon command to gather additional metrics
-metrics-port int
      Metrics server port (default 8870)
-exporter string
      Metrics exporter type (prometheus, otlphttp, or otlpgrpc) (default "prometheus")
-log-level string
      Log level (debug, info, warn, error) (default "info")
-log-format string
      Log format (text or json) (default "text")
```

### Environment variables

The following standard OpenTelemetry environment variables are supported:

- `OTEL_EXPORTER_OTLP_CERTIFICATE`: certificate file path for TLS verification
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint URL (e.g. `http://localhost:4318` for HTTP, `localhost:4317` for gRPC)
- `OTEL_EXPORTER_OTLP_HEADERS`: headers to include in requests (comma-separated key=value pairs)
- `OTEL_EXPORTER_OTLP_INSECURE`: use insecure transport for gRPC (default false)
- `OTEL_EXPORTER_OTLP_TIMEOUT`: timeout for OTLP export requests in milliseconds (default 10000)
- `OTEL_METRIC_EXPORT_INTERVAL`: metric export interval in milliseconds (default 60000)
- `OTEL_METRIC_EXPORT_TIMEOUT`: metric export timeout in milliseconds (default 30000)
- `OTEL_RESOURCE_ATTRIBUTES`: extra resource attributes (comma-separated key=value pairs)
- `OTEL_SERVICE_NAME`: service name resource attribute (default "jka-server")

### Metrics

The following base metrics are exposed:

- `jka.clients.connected`: current number of clients connected
- `jka.clients.limit`: maximum number of clients allowed
- `jka.clients.ping`: player ping in milliseconds (with player name label)

The following RPMod metrics are exposed when the `-enable-rpmetrics` flag is set:

- `jka.server.cs.limit`: maximum number of Config String characters allowed
- `jka.server.cs.usage`: current number of Config String characters used
- `jka.server.entities.limit`: maximum number of entities allowed
- `jka.server.entities.usage`: current number of entities used
- `jka.server.rpcs.limit`: maximum number of RPCS characters allowed
- `jka.server.rpcs.usage`: current number of RPCS characters used
- `jka.server.uptime`: server uptime in milliseconds

When using the Prometheus exporter (default), metrics are available at `http://localhost:8870/metrics`. Note that dots (`.`) are replaced with underscores (`_`) in Prometheus metric names.

Additionally, a `http://localhost:8870/health` endpoint is available for health checking, e.g. by Kubernetes liveness/readiness probes.

## Building from source

1. Clone the repository:
   ```bash
   git clone https://github.com/fcrespel/jka-exporter.git
   cd jka-exporter
   ```

2. Build the project:
   ```bash
   go build
   ```

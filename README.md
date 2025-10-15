# JKA Exporter

A metrics exporter for Star Wars Jedi Knight: Jedi Academy dedicated servers, supporting Prometheus and OTLP formats.

## Usage

The following commands can be used to run the exporter with [Docker](https://docs.docker.com/engine/).

```bash
# Show syntax help
docker run --rm ghcr.io/fcrespel/jka-exporter:master -help

# Start in the background with Prometheus exporter (default)
docker run -d --name jka-exporter -p 8870:8870 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070

# Start in the background with OTLP HTTP exporter
docker run -d --name jka-exporter -p 8870:8870 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070 -exporter otlphttp -otlp-endpoint otlp-receiver:4318

# Start in the background with OTLP GRPC exporter
docker run -d --name jka-exporter -p 8870:8870 ghcr.io/fcrespel/jka-exporter:master -host <JKA server host or IP> -port 29070 -exporter otlpgrpc -otlp-endpoint otlp-receiver:4317

# Stop exporter
docker stop jka-exporter

# Delete container
docker rm jka-exporter
```

### Options

```
Server options:
  -host string
        Server host name or IP address (default "localhost")
  -port int
        Server port (default 29070)

Metrics options:
  -metrics-port int
        Metrics server port (default 8870)
  -exporter string
        Metrics exporter type (prometheus, otlphttp, or otlpgrpc) (default "prometheus")
  -otlp-endpoint string
        OTLP endpoint (default "localhost:4318")
  -otlp-timeout duration
        OTLP request timeout (default 10s)
  -otlp-interval duration
        OTLP metric collection interval (default 60s)

Logging options:
  -log-level string
        Log level (debug, info, warn, error) (default "info")
  -log-format string
        Log format (text or json) (default "text")
```

### Metrics

The following metrics are exposed:

- `jka.clients.connected`: Current number of clients connected
- `jka.clients.max`: Maximum number of clients allowed
- `jka.clients.ping`: Player ping in milliseconds (with player name label)

When using the Prometheus exporter (default), metrics are available at `http://localhost:8870/metrics`.

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

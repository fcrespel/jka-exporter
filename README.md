# JKA Exporter

Jedi Academy metrics exporter supporting Prometheus and OTLP formats.

## Installation

```bash
go install github.com/fcrespel/jka-exporter@latest
```

## Usage

```bash
jka-exporter [options]
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

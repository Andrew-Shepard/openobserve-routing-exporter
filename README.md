# openobserve-routing-exporter

Custom OTel Collector exporter for OpenObserve. Routes logs to streams based on resource attributes.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [OTel Collector Builder (ocb)](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)

Install ocb:
```sh
go install go.opentelemetry.io/collector/cmd/builder@v0.153.0
```

## Build

```sh
builder --config otelcol-builder.yaml
```

Output binary: `./dist/otelcol-openobserve`

> **Windows:** If `go` is not on your PATH, set it in `otelcol-builder.yaml` under `dist.go`.

## Run

```sh
./dist/otelcol-openobserve --config collector.yaml
```

## Exporter config

```yaml
exporters:
  openobserverouting:
    endpoint: "https://openobserve.example.com/api/myorg/"
    headers:
      Authorization: "Basic <token>"
    default_stream: "default"
    rules:
      - attribute: k8s.namespace.name
        equals: payments
        stream: payments-logs
      - attribute: k8s.namespace.name
        stream: "${k8s.namespace.name}-logs"
```

Routing priority per log batch:
1. `openobserve.stream` resource attribute — used directly if present
2. First matching rule
3. `default_stream`

# openobserve-routing-exporter

Custom OTel Collector exporter targeting OpenObserve via OTLP/HTTP.
Sets `stream-name` header per request based on config-driven routing rules.

## Routing rules
- Ordered, first match wins, fallback to default stream
- Match on: attribute exists, or attribute equals value  
- Dynamic stream names via `${attribute}` interpolation
- Operates on resource attributes

## Reference
- exporter/otlphttpexporter — HTTP sending mechanics
- exporter/routingconnector — routing config pattern
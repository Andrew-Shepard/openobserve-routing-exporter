package openobserveroutingexporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var interpolationRe = regexp.MustCompile(`\$\{([^}]+)\}`)

type routingExporter struct {
	cfg    *Config
	client *http.Client
}

func newExporter(cfg *Config, _ component.TelemetrySettings) (*routingExporter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &routingExporter{cfg: cfg}, nil
}

func (e *routingExporter) start(_ context.Context, _ component.Host) error {
	e.client = &http.Client{}
	return nil
}

func (e *routingExporter) shutdown(_ context.Context) error {
	if e.client != nil {
		e.client.CloseIdleConnections()
	}
	return nil
}

// --- routing helpers ---

// matchStream returns the target stream name for a set of resource attributes.
// Priority: openobserve.stream attribute → first matching rule → default stream.
func (e *routingExporter) matchStream(attrs pcommon.Map) string {
	if v, ok := attrs.Get("openobserve.stream"); ok {
		return v.AsString()
	}
	for _, rule := range e.cfg.Rules {
		val, exists := attrs.Get(rule.Attribute)
		if !exists {
			continue
		}
		if rule.Equals != "" && val.AsString() != rule.Equals {
			continue
		}
		return interpolate(rule.Stream, attrs)
	}
	return e.cfg.DefaultStream
}

// interpolate replaces ${attr.name} placeholders with attribute values.
func interpolate(stream string, attrs pcommon.Map) string {
	return interpolationRe.ReplaceAllStringFunc(stream, func(match string) string {
		key := match[2 : len(match)-1] // strip ${ and }
		if v, ok := attrs.Get(key); ok {
			return v.AsString()
		}
		return match // no attribute found — leave placeholder as-is
	})
}

// --- HTTP send ---

func (e *routingExporter) send(ctx context.Context, path, stream string, body []byte) error {
	url := strings.TrimRight(e.cfg.Endpoint, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("stream-name", stream)
	for k, v := range e.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// --- push functions ---

func (e *routingExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	if ld.ResourceLogs().Len() == 0 {
		return nil
	}

	// Fast path: if all ResourceLogs route to the same stream, marshal directly.
	firstStream := e.matchStream(ld.ResourceLogs().At(0).Resource().Attributes())
	allSame := true
	for i := 1; i < ld.ResourceLogs().Len(); i++ {
		if e.matchStream(ld.ResourceLogs().At(i).Resource().Attributes()) != firstStream {
			allSame = false
			break
		}
	}
	if allSame {
		b, err := (&plog.ProtoMarshaler{}).MarshalLogs(ld)
		if err != nil {
			return err
		}
		return e.send(ctx, "/v1/logs", firstStream, b)
	}

	// Slow path: split by stream, copy each ResourceLogs into its group.
	groups := map[string]plog.Logs{}
	for i := range ld.ResourceLogs().Len() {
		rl := ld.ResourceLogs().At(i)
		stream := e.matchStream(rl.Resource().Attributes())
		if _, ok := groups[stream]; !ok {
			groups[stream] = plog.NewLogs()
		}
		rl.CopyTo(groups[stream].ResourceLogs().AppendEmpty())
	}

	m := &plog.ProtoMarshaler{}
	for stream, logs := range groups {
		b, err := m.MarshalLogs(logs)
		if err != nil {
			return err
		}
		if err := e.send(ctx, "/v1/logs", stream, b); err != nil {
			return err
		}
	}
	return nil
}

func (e *routingExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	b, err := (&ptrace.ProtoMarshaler{}).MarshalTraces(td)
	if err != nil {
		return err
	}
	return e.send(ctx, "/v1/traces", e.cfg.DefaultStream, b)
}

func (e *routingExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	b, err := (&pmetric.ProtoMarshaler{}).MarshalMetrics(md)
	if err != nil {
		return err
	}
	return e.send(ctx, "/v1/metrics", e.cfg.DefaultStream, b)
}

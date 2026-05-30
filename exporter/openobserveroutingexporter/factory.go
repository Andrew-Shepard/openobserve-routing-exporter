package openobserveroutingexporter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const typeStr = "openobserverouting"

// NewFactory creates the component factory registered with the Collector.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, component.StabilityLevelDevelopment),
		exporter.WithLogs(createLogsExporter, component.StabilityLevelDevelopment),
		exporter.WithMetrics(createMetricsExporter, component.StabilityLevelDevelopment),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		DefaultStream: "default",
		Headers:       map[string]string{},
		Rules:         []RoutingRule{},
		RetryConfig:   configretry.NewDefaultBackOffConfig(),
	}
}

func createTracesExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Traces, error) {
	c := cfg.(*Config)
	exp, err := newExporter(c, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewTraces(ctx, set, cfg,
		exp.pushTraces,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(c.RetryConfig),
	)
}

func createLogsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Logs, error) {
	c := cfg.(*Config)
	exp, err := newExporter(c, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewLogs(ctx, set, cfg,
		exp.pushLogs,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(c.RetryConfig),
	)
}

func createMetricsExporter(ctx context.Context, set exporter.Settings, cfg component.Config) (exporter.Metrics, error) {
	c := cfg.(*Config)
	exp, err := newExporter(c, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewMetrics(ctx, set, cfg,
		exp.pushMetrics,
		exporterhelper.WithStart(exp.start),
		exporterhelper.WithShutdown(exp.shutdown),
		exporterhelper.WithRetry(c.RetryConfig),
	)
}

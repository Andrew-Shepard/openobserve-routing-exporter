package openobserveroutingexporter

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/config/configretry"
)

// Config defines the configuration for the OpenObserve routing exporter.
type Config struct {
	// Endpoint is the OpenObserve OTLP/HTTP base URL.
	Endpoint string `mapstructure:"endpoint"`

	// Headers are additional HTTP headers sent with every request.
	Headers map[string]string `mapstructure:"headers"`

	// DefaultStream is used when no routing rule matches.
	DefaultStream string `mapstructure:"default_stream"`

	// Rules is the ordered list of routing rules. First match wins.
	Rules []RoutingRule `mapstructure:"rules"`

	// RetryConfig configures exponential backoff retry on failed requests.
	RetryConfig configretry.BackOffConfig `mapstructure:"retry_on_failure"`
}

// RoutingRule defines a single routing rule evaluated against resource attributes.
type RoutingRule struct {
	// Attribute is the resource attribute key to match on.
	Attribute string `mapstructure:"attribute"`

	// Equals is optional. Omitting it matches any value (existence check).
	Equals string `mapstructure:"equals"`

	// Stream is the target stream name. Supports ${attribute} interpolation.
	Stream string `mapstructure:"stream"`
}

// Validate is called by the Collector on startup to catch config errors early.
func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.DefaultStream == "" {
		return errors.New("default_stream is required")
	}
	for i, rule := range c.Rules {
		if rule.Attribute == "" {
			return fmt.Errorf("rules[%d]: attribute is required", i)
		}
		if rule.Stream == "" {
			return fmt.Errorf("rules[%d]: stream is required", i)
		}
	}
	return nil
}

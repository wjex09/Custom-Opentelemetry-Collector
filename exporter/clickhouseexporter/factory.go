// exporter/clickhouseexporter/factory.go
package clickhouseexporter

import (
    "context"
    "fmt"

    "go.opentelemetry.io/collector/component"
    "go.opentelemetry.io/collector/exporter"
)

const typeStr = "clickhouse"

func NewFactory() exporter.Factory {
    return exporter.NewFactory(
        typeStr,
        createDefaultConfig,
        exporter.WithMetrics(createMetricsExporter, component.StabilityLevelBeta),
    )
}

func createDefaultConfig() component.Config {
    return &Config{
        Endpoint: "t3v0qmphlz.ap-south-1.aws.clickhouse.cloud:8443",
        Username: "default",
        Password: "CSgSGq25q4Re.",
        Database: "otel",
        Secure:   true,
    }
}

func createMetricsExporter(
    ctx context.Context,
    params exporter.CreateSettings,
    cfg component.Config,
) (exporter.Metrics, error) {
    oCfg := cfg.(*Config)
    if oCfg.Endpoint == "" {
        return nil, fmt.Errorf("endpoint must be specified")
    }

    return NewClickHouseExporter(ctx)
}
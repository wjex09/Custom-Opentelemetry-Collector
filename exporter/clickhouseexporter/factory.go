// exporter/clickhouseexporter/factory.go
package clickhouseexporter

import (
    "context"
    "fmt"
    "os"

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
        Endpoint: os.Getenv("CLICKHOUSE_ENDPOINT"),
        Username: os.Getenv("CLICKHOUSE_USERNAME"),
        Password: os.Getenv("CLICKHOUSE_PASSWORD"),
        Database: os.Getenv("CLICKHOUSE_DATABASE"),
        Secure:   true,
    }
}

func createMetricsExporter(
    ctx context.Context,
    params exporter.CreateSettings,
    cfg component.Config,
) (exporter.Metrics, error) {
    oCfg := cfg.(*Config)
    
    // Validate required environment variables
    if oCfg.Endpoint == "" {
        oCfg.Endpoint = os.Getenv("CLICKHOUSE_ENDPOINT")
    }
    if oCfg.Username == "" {
        oCfg.Username = os.Getenv("CLICKHOUSE_USERNAME")
    }
    if oCfg.Password == "" {
        oCfg.Password = os.Getenv("CLICKHOUSE_PASSWORD")
    }
    if oCfg.Database == "" {
        oCfg.Database = os.Getenv("CLICKHOUSE_DATABASE")
    }

    // Final validation
    if oCfg.Endpoint == "" {
        return nil, fmt.Errorf("endpoint must be specified via config or CLICKHOUSE_ENDPOINT environment variable")
    }
    if oCfg.Username == "" {
        return nil, fmt.Errorf("username must be specified via config or CLICKHOUSE_USERNAME environment variable")
    }
    if oCfg.Password == "" {
        return nil, fmt.Errorf("password must be specified via config or CLICKHOUSE_PASSWORD environment variable")
    }
    if oCfg.Database == "" {
        return nil, fmt.Errorf("database must be specified via config or CLICKHOUSE_DATABASE environment variable")
    }

    return NewClickHouseExporter(ctx)
}

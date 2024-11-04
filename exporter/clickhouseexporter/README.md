# OpenTelemetry ClickHouse Exporter

A custom OpenTelemetry exporter that sends metrics to ClickHouse. This exporter handles metric data collection and storage in ClickHouse for efficient time series analytics.

## Structure

```
exporter/clickhouseexporter/
├── exporter.go       # Main exporter implementation
├── config.go         # Configuration definitions
├── factory.go        # Factory methods for collector
├── examples/
│   └── main.go      # Example usage
└── go.mod
```

## Components

### exporter.go
Contains the main exporter implementation that processes and sends OpenTelemetry metrics to ClickHouse.

```go
// Key functionalities:
- ConsumeMetrics: Processes incoming metrics
- exportDataPoints: Handles metric data point export
- Shutdown: Cleanup resources
```

### config.go
Defines the configuration structure for the exporter.

```go
type Config struct {
    Endpoint string  // ClickHouse endpoint
    Username string  // Database username
    Password string  // Database password
    Database string  // Database name
    Secure   bool    // Use secure connection
}
```

### factory.go
Factory methods for creating and configuring the exporter within the OpenTelemetry collector.

```go
// Key functions:
- NewFactory: Creates a new exporter factory
- createDefaultConfig: Provides default configuration
- createMetricsExporter: Creates metrics exporter instance
```

### examples/main.go
Example implementation showing how to use the exporter directly.

## Setup

1. Add the exporter to your project:
```bash
go get github.com/open-telemetry/opentelemetry-collector-contrib/exporter/clickhouseexporter
```

2. Configure ClickHouse connection:
```go
config := &Config{
    Endpoint: "your-clickhouse-host:9440",
    Username: "default",
    Password: "your-password",
    Database: "otel",
    Secure: true,
}
```

## Usage

### Direct Usage
```go
package main

import (
    "context"
    "log"
    "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/clickhouseexporter"
)

func main() {
    ctx := context.Background()
    
    // Create exporter
    exp, err := clickhouseexporter.NewClickHouseExporter(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer exp.Shutdown(ctx)

    // Create metrics
    metrics := createSampleMetrics()

    // Export metrics
    if err := exp.ConsumeMetrics(ctx, metrics); err != nil {
        log.Printf("Failed to export metrics: %v", err)
    }
}
```

### As Collector Component
```yaml
exporters:
  clickhouse:
    endpoint: your-clickhouse-host:9440
    username: default
    password: your-password
    database: otel
    secure: true

service:
  pipelines:
    metrics:
      receivers: [your-receivers]
      processors: [your-processors]
      exporters: [clickhouse]
```

## ClickHouse Schema

The exporter expects the following table structure:

```sql
CREATE TABLE IF NOT EXISTS otel.metrics
(
    timestamp DateTime64(9),
    metric_name LowCardinality(String),
    metric_type Enum8('gauge' = 1, 'sum' = 2, 'histogram' = 3),
    value Float64,
    labels Map(LowCardinality(String), String),
    service_name LowCardinality(String),
    host_name LowCardinality(String)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (metric_name, timestamp, service_name)
TTL toDateTime(timestamp) + INTERVAL 30 DAY;
```

## Metrics Support

The exporter handles the following metric types:
- Gauge metrics
- Sum metrics
- Histogram metrics

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| endpoint | ClickHouse server endpoint | required |
| username | Database username | required |
| password | Database password | required |
| database | Database name | "otel" |
| secure | Use TLS connection | true |

## Example Metrics

```go
func createSampleMetrics() pmetric.Metrics {
    metrics := pmetric.NewMetrics()
    rm := metrics.ResourceMetrics().AppendEmpty()
    
    // Add resource attributes
    attributes := rm.Resource().Attributes()
    attributes.PutStr("service.name", "test-service")
    attributes.PutStr("host.name", "test-host")

    sm := rm.ScopeMetrics().AppendEmpty()
    metric := sm.Metrics().AppendEmpty()
    
    // Add gauge metric
    metric.SetName("system.memory.usage")
    gauge := metric.SetEmptyGauge()
    dp := gauge.DataPoints().AppendEmpty()
    dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
    dp.SetDoubleValue(8589934592)

    return metrics
}
```

## Error Handling

The exporter implements comprehensive error handling:
- Connection errors
- Data validation errors
- Export failures
- Resource cleanup

## Performance Considerations

- Uses batch inserts for better performance
- Implements connection pooling
- Handles large metric volumes efficiently
- Supports data compression

## Development

### Prerequisites
- Go 1.21 or later
- ClickHouse server
- OpenTelemetry Collector development environment

### Testing
```bash
go test ./...
```

### Building
```bash
go build ./...
```

## Troubleshooting

Common issues:

1. Connection Failures
   - Verify ClickHouse credentials
   - Check network connectivity
   - Verify SSL/TLS settings

2. Data Export Issues
   - Check table schema
   - Verify metric types
   - Check data format

3. Performance Issues
   - Monitor batch sizes
   - Check ClickHouse server resources
   - Review query performance


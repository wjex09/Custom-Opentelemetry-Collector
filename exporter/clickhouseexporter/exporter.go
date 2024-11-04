// exporter/clickhouseexporter/exporter.go
package clickhouseexporter

import (
    "context"
    "fmt"
    "time"
    "crypto/tls"
    "database/sql"

    "github.com/ClickHouse/clickhouse-go/v2"
    "go.opentelemetry.io/collector/component"
    "go.opentelemetry.io/collector/consumer"
    "go.opentelemetry.io/collector/exporter"
    "go.opentelemetry.io/collector/pdata/pcommon"
    "go.opentelemetry.io/collector/pdata/pmetric"
    "go.uber.org/zap"
)

type clickhouseExporter struct {
    cfg    *Config
    db     *sql.DB
    logger *zap.Logger
}

// NewClickHouseExporter creates a new instance of clickhouseExporter for standalone usage
func NewClickHouseExporter(ctx context.Context) (exporter.Metrics, error) {
    logger, _ := zap.NewProduction()

    // Connect using the native protocol
    db := clickhouse.OpenDB(&clickhouse.Options{
        Addr: []string{"t3v0qmphlz.ap-south-1.aws.clickhouse.cloud:9440"},
        Protocol: clickhouse.Native,
        TLS: &tls.Config{},
        Auth: clickhouse.Auth{
            Database: "otel",
            Username: "default",
            Password: "CSgSGq25q4Re.",
        },
        Debug: true,
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
        Compression: &clickhouse.Compression{
            Method: clickhouse.CompressionLZ4,
        },
    })

    // Test the connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
    }

    return &clickhouseExporter{
        db:     db,
        logger: logger,
    }, nil
}

// Capabilities implements the consumer.Capabilities interface.
func (e *clickhouseExporter) Capabilities() consumer.Capabilities {
    return consumer.Capabilities{MutatesData: false}
}

// Start implements the component.Component interface.
func (e *clickhouseExporter) Start(_ context.Context, _ component.Host) error {
    return nil
}

func (e *clickhouseExporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
    metrics := md.ResourceMetrics()

    // Prepare the query
    stmt, err := e.db.Prepare(`
        INSERT INTO otel.metrics (
            timestamp,
            metric_name,
            metric_type,
            value,
            labels,
            service_name,
            host_name
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    for i := 0; i < metrics.Len(); i++ {
        rm := metrics.At(i)
        resource := rm.Resource()
        
        var serviceName, hostName string
        if serviceAttr, ok := resource.Attributes().Get("service.name"); ok {
            serviceName = serviceAttr.Str()
        }
        if hostAttr, ok := resource.Attributes().Get("host.name"); ok {
            hostName = hostAttr.Str()
        }

        ilm := rm.ScopeMetrics()
        for j := 0; j < ilm.Len(); j++ {
            ilMetrics := ilm.At(j).Metrics()
            
            for k := 0; k < ilMetrics.Len(); k++ {
                metric := ilMetrics.At(k)
                
                switch metric.Type() {
                case pmetric.MetricTypeGauge:
                    if err := e.exportDataPoints(ctx, stmt, metric.Gauge().DataPoints(), metric.Name(), "gauge", serviceName, hostName); err != nil {
                        e.logger.Error("Failed to export gauge metric", zap.Error(err))
                    }

                case pmetric.MetricTypeSum:
                    if err := e.exportDataPoints(ctx, stmt, metric.Sum().DataPoints(), metric.Name(), "sum", serviceName, hostName); err != nil {
                        e.logger.Error("Failed to export sum metric", zap.Error(err))
                    }

                case pmetric.MetricTypeHistogram:
                    dp := metric.Histogram().DataPoints()
                    for l := 0; l < dp.Len(); l++ {
                        point := dp.At(l)
                        labels := attributesToMap(point.Attributes())
                        
                        if point.Count() > 0 {
                            value := point.Sum() / float64(point.Count())
                            
                            _, err := stmt.ExecContext(ctx,
                                time.Unix(0, int64(point.Timestamp())).UTC(),
                                metric.Name(),
                                "histogram",
                                value,
                                labels,
                                serviceName,
                                hostName,
                            )
                            if err != nil {
                                e.logger.Error("Failed to insert histogram metric",
                                    zap.Error(err),
                                    zap.String("metric", metric.Name()),
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    return nil
}

func (e *clickhouseExporter) exportDataPoints(
    ctx context.Context,
    stmt *sql.Stmt,
    dp pmetric.NumberDataPointSlice,
    metricName string,
    metricType string,
    serviceName string,
    hostName string,
) error {
    for i := 0; i < dp.Len(); i++ {
        point := dp.At(i)
        labels := attributesToMap(point.Attributes())
        
        var value float64
        switch point.ValueType() {
        case pmetric.NumberDataPointValueTypeDouble:
            value = point.DoubleValue()
        case pmetric.NumberDataPointValueTypeInt:
            value = float64(point.IntValue())
        }

        _, err := stmt.ExecContext(ctx,
            time.Unix(0, int64(point.Timestamp())).UTC(),
            metricName,
            metricType,
            value,
            labels,
            serviceName,
            hostName,
        )
        if err != nil {
            return fmt.Errorf("failed to insert metric %s: %w", metricName, err)
        }
    }
    return nil
}

func attributesToMap(attrs pcommon.Map) map[string]string {
    result := make(map[string]string)
    attrs.Range(func(k string, v pcommon.Value) bool {
        result[k] = v.AsString()
        return true
    })
    return result
}

func (e *clickhouseExporter) Shutdown(ctx context.Context) error {
    if e.db != nil {
        return e.db.Close()
    }
    return nil
}
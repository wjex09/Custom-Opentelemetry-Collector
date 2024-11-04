// exporter/clickhouseexporter/examples/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "crypto/tls"
    "os"

    "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/clickhouseexporter"
    "go.opentelemetry.io/collector/pdata/pcommon"
    "go.opentelemetry.io/collector/pdata/pmetric" 
    "github.com/joho/godotenv" 
    "path/filepath"
)



// exporter/clickhouseexporter/examples/main.go
func init() {
    // Print working directory
    workDir, err := os.Getwd()
    if err != nil {
        log.Printf("Error getting working directory: %v", err)
    }
    fmt.Printf("Working directory: %s\n", workDir)

    // Try multiple possible locations
    possiblePaths := []string{
        "../.env",                // Relative to examples
        "../../clickhouseexporter/.env", // From examples to clickhouseexporter
        ".env",                   // Current directory
    }

    var loaded bool
    for _, path := range possiblePaths {
        absPath, err := filepath.Abs(path)
        if err != nil {
            continue
        }
        fmt.Printf("Trying to load .env from: %s\n", absPath)
        
        if err := godotenv.Load(path); err == nil {
            fmt.Printf("Successfully loaded .env from: %s\n", absPath)
            loaded = true
            break
        }
    }

    if !loaded {
        log.Printf("Could not find .env file in any location")
    }

    // Verify environment variables
    vars := []string{"CLICKHOUSE_ENDPOINT", "CLICKHOUSE_USERNAME", "CLICKHOUSE_PASSWORD", "CLICKHOUSE_DATABASE"}
    for _, v := range vars {
        if val := os.Getenv(v); val != "" {
            fmt.Printf("%s is set\n", v)
        } else {
            fmt.Printf("%s is NOT set\n", v)
        }
    }
}
func main() {
    ctx := context.Background()
    
    // Create the exporter
    exp, err := clickhouseexporter.NewClickHouseExporter(ctx)
    if err != nil {
        log.Fatalf("Failed to create exporter: %v", err)
    }

    // Create sample metrics
    metrics := createSampleMetrics()
    fmt.Printf("Created sample metrics with %d resource metrics\n", 
        metrics.ResourceMetrics().Len())

    // Export metrics
    if err := exp.ConsumeMetrics(ctx, metrics); err != nil {
        log.Fatalf("Failed to export metrics: %v", err)
    }
    fmt.Println("Successfully exported metrics")

    // Add a delay to ensure metrics are written
    time.Sleep(time.Second * 2)

    // Get configuration from environment variables
    endpoint := os.Getenv("CLICKHOUSE_ENDPOINT")
    username := os.Getenv("CLICKHOUSE_USERNAME")
    password := os.Getenv("CLICKHOUSE_PASSWORD")
    database := os.Getenv("CLICKHOUSE_DATABASE")

    // Validate required environment variables
    if endpoint == "" || username == "" || password == "" || database == "" {
        log.Fatal("Missing required environment variables")
    }

    // Verify the inserted data
    conn := clickhouse.OpenDB(&clickhouse.Options{
        Addr: []string{endpoint},
        Protocol: clickhouse.Native,
        TLS: &tls.Config{},
        Auth: clickhouse.Auth{
            Database: database,
            Username: username,
            Password: password,
        },
    })

    // First, verify the table exists
    tableCheckQuery := fmt.Sprintf(`
        SELECT count()
        FROM system.tables
        WHERE database = '%s' AND name = 'metrics'
    `, database)
    
    var tableCount int
    if err := conn.QueryRow(tableCheckQuery).Scan(&tableCount); err != nil {
        log.Fatalf("Failed to check table existence: %v", err)
    }
    if tableCount == 0 {
        log.Fatalf("Table '%s.metrics' does not exist", database)
    }

    // Check if any data exists
    countQuery := fmt.Sprintf(`
        SELECT count()
        FROM %s.metrics
    `, database)
    
    var recordCount int
    if err := conn.QueryRow(countQuery).Scan(&recordCount); err != nil {
        log.Fatalf("Failed to count records: %v", err)
    }
    fmt.Printf("\nTotal records in table: %d\n", recordCount)

    // Query the last 5 metrics with a more lenient filter
    query := fmt.Sprintf(`
        SELECT 
            timestamp,
            metric_name,
            metric_type,
            value,
            labels,
            service_name,
            host_name
        FROM %s.metrics
        ORDER BY timestamp DESC 
        LIMIT 15
    `, database)

    rows, err := conn.Query(query)
    if err != nil {
        log.Fatalf("Failed to query metrics: %v", err)
    }
    defer rows.Close()

    fmt.Println("\nLast 15 metrics in the database:")
    fmt.Println("--------------------------------------------------")
    for rows.Next() {
        var (
            ts          time.Time
            metricName  string
            metricType  string
            value       float64
            labels      map[string]string
            serviceName string
            hostName    string
        )
        if err := rows.Scan(&ts, &metricName, &metricType, &value, &labels, &serviceName, &hostName); err != nil {
            log.Printf("Error scanning row: %v", err)
            continue
        }
        fmt.Printf("Time: %v\nMetric: %s\nType: %s\nValue: %f\nLabels: %v\nService: %s\nHost: %s\n\n",
            ts, metricName, metricType, value, labels, serviceName, hostName)
    }

    if err := rows.Err(); err != nil {
        log.Printf("Error iterating rows: %v", err)
    }
}


func createSampleMetrics() pmetric.Metrics {
    metrics := pmetric.NewMetrics()
    rm := metrics.ResourceMetrics().AppendEmpty()
    
    // Add resource attributes
    attributes := rm.Resource().Attributes()
    attributes.PutStr("service.name", "test-service")
    attributes.PutStr("host.name", "test-host")

    sm := rm.ScopeMetrics().AppendEmpty()
    
    // Add a gauge metric
    addGaugeMetric(sm, "system.memory.usage", 8589934592)
    
    // Add a counter metric
    addCounterMetric(sm, "system.cpu.time", 12.7)
    
    // Add a metric with labels
    addLabeledMetric(sm, "http.requests", 42, map[string]string{
        "method": "GET",
        "path": "/api/users",
        "status": "200",
    })

    return metrics
}

func addGaugeMetric(sm pmetric.ScopeMetrics, name string, value float64) {
    metric := sm.Metrics().AppendEmpty()
    metric.SetName(name)
    gauge := metric.SetEmptyGauge()
    dp := gauge.DataPoints().AppendEmpty()
    dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
    dp.SetDoubleValue(value)
}

func addCounterMetric(sm pmetric.ScopeMetrics, name string, value float64) {
    metric := sm.Metrics().AppendEmpty()
    metric.SetName(name)
    sum := metric.SetEmptySum()
    sum.SetIsMonotonic(true)
    dp := sum.DataPoints().AppendEmpty()
    dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
    dp.SetDoubleValue(value)
}

func addLabeledMetric(sm pmetric.ScopeMetrics, name string, value float64, labels map[string]string) {
    metric := sm.Metrics().AppendEmpty()
    metric.SetName(name)
    gauge := metric.SetEmptyGauge()
    dp := gauge.DataPoints().AppendEmpty()
    dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
    dp.SetDoubleValue(value)
    
    // Add labels
    for k, v := range labels {
        dp.Attributes().PutStr(k, v)
    }
}
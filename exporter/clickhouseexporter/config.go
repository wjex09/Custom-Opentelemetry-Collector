// config.go
package clickhouseexporter

import (
    "os"
    "fmt"
)

type Config struct {
    Endpoint string
    Username string
    Password string
    Database string
    Secure   bool
}

func NewConfig() (*Config, error) {
    password := os.Getenv("CLICKHOUSE_PASSWORD")
    if password == "" {
        return nil, fmt.Errorf("CLICKHOUSE_PASSWORD environment variable is not set")
    }

    return &Config{
        Endpoint: getEnvWithDefault("CLICKHOUSE_ENDPOINT", "t3v0qmphlz.ap-south-1.aws.clickhouse.cloud:9440"),
        Username: getEnvWithDefault("CLICKHOUSE_USERNAME", "default"),
        Password: password,
        Database: getEnvWithDefault("CLICKHOUSE_DATABASE", "otel"),
        Secure:   true,
    }, nil
}

func getEnvWithDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
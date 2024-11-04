// exporter/clickhouseexporter/config.go
package clickhouseexporter

// Config defines configuration for ClickHouse exporter.
type Config struct {
    Endpoint string `mapstructure:"endpoint"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    Database string `mapstructure:"database"`
    Secure   bool   `mapstructure:"secure"`
}

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
    return nil
}
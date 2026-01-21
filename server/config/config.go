package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Filter   FilterConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Path string
}

type FilterConfig struct {
	APIURL         string
	MaxConcurrent  int
	ResultChanSize int
}

// Global config instance
var AppConfig *Config

// Load reads configuration from file, environment variables, and defaults
func Load() *Config {
	// Set config file name and paths to search
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("database.path", "smsrequest.db")
	viper.SetDefault("filter.apiurl", "http://192.168.8.100:8000/api/v0/filter/sms")
	viper.SetDefault("filter.maxconcurrent", 5)
	viper.SetDefault("filter.resultchansize", 10)

	// Read config file (optional - won't error if not found)
	viper.ReadInConfig()

	// Enable environment variable reading
	viper.SetEnvPrefix("MICROSMS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Unmarshal into struct
	AppConfig = &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
			Host: viper.GetString("server.host"),
		},
		Database: DatabaseConfig{
			Path: viper.GetString("database.path"),
		},
		Filter: FilterConfig{
			APIURL:         viper.GetString("filter.apiurl"),
			MaxConcurrent:  viper.GetInt("filter.maxconcurrent"),
			ResultChanSize: viper.GetInt("filter.resultchansize"),
		},
	}

	return AppConfig
}

// Print displays the current configuration
func (c *Config) Print() {
	fmt.Println("=== Application Configuration ===")
	fmt.Printf("Server Address: %s:%s\n", c.Server.Host, c.Server.Port)
	fmt.Printf("Database Path: %s\n", c.Database.Path)
	fmt.Printf("Filter API URL: %s\n", c.Filter.APIURL)
	fmt.Printf("Filter Max Concurrent: %d\n", c.Filter.MaxConcurrent)
	fmt.Printf("Filter Result Channel Size: %d\n", c.Filter.ResultChanSize)
	fmt.Println("=================================")
}

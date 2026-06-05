package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost            string        `mapstructure:"db_host"`
	DBPort            int           `mapstructure:"db_port"`
	DBUser            string        `mapstructure:"db_user"`
	DBPassword        string        `mapstructure:"db_password"`
	DBName            string        `mapstructure:"db_name"`
	DBSSLMode         string        `mapstructure:"db_sslmode"`
	DBMaxConns        int32         `mapstructure:"db_max_conns"`
	DBMinConns        int32         `mapstructure:"db_min_conns"`
	DBMaxConnLifetime time.Duration `mapstructure:"db_max_conn_lifetime"`
	DBMaxIdleTime     time.Duration `mapstructure:"db_max_idle_time"`
	Port              int           `mapstructure:"port"`
	RedisAddr         string        `mapstructure:"redis_addr"`
	RedisPass         string        `mapstructure:"redis_pass"`
	RedisDb           int           `mapstructure:"redis_db"`
	EventSvcUrl       string        `mapstructure:"event_svc_url"`

	AiMaxTokens int `mapstructure:"ai_max_tokens"`

	ClaudeModel  string
	ClaudeApiKey string

	GeminiApiKey string `mapstructure:"gemini_api_key"`
	GeminiModel  string `mapstructure:"gemini_model"`

	AethexBaseUrl string  `mapstructure:"aethex_base_url"`
	AethexApiKey  string  `mapstructure:"aethex_api_key"`
	AethexAgentId *string `mapstructure:"aethex_agent_id"`
}

func LoadConfig() Config {
	fmt.Println("Load Config")

	v := viper.New()

	// ---------------------------
	// Read .env file
	// ---------------------------
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("/root")

	// ---------------------------
	// Enable reading environment variables
	// ---------------------------
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ---------------------------
	// Set defaults matching your .env
	// ---------------------------
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "macbookpro")
	v.SetDefault("DB_PASSWORD", "")
	v.SetDefault("DB_NAME", "pazar360")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_CONNS", 20)
	v.SetDefault("DB_MIN_CONNS", 2)
	v.SetDefault("DB_MAX_CONN_LIFETIME", "1h")
	v.SetDefault("DB_MAX_IDLE_TIME", "30m")
	v.SetDefault("PORT", 8081)

	// ---------------------------
	// Read .env file if present
	// ---------------------------
	if err := v.ReadInConfig(); err == nil {
		log.Println("Loaded .env file:", v.ConfigFileUsed())
	} else {
		log.Println("No .env file found, using environment variables")
	}

	// Parse duration strings
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}

	// Manually parse durations since viper doesn't always handle them correctly
	if maxLifetime := v.GetString("DB_MAX_CONN_LIFETIME"); maxLifetime != "" {
		duration, err := time.ParseDuration(maxLifetime)
		if err == nil {
			cfg.DBMaxConnLifetime = duration
		}
	}

	if maxIdleTime := v.GetString("DB_MAX_IDLE_TIME"); maxIdleTime != "" {
		duration, err := time.ParseDuration(maxIdleTime)
		if err == nil {
			cfg.DBMaxIdleTime = duration
		}
	}

	return cfg
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// GetDBPoolConfig returns pgxpool configuration
func (c *Config) GetDBPoolConfig() (int32, int32, time.Duration, time.Duration) {
	return c.DBMaxConns, c.DBMinConns, c.DBMaxConnLifetime, c.DBMaxIdleTime
}

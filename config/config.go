// config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/joho/godotenv"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    CORS     CORSConfig
    Logger   LoggerConfig
    Redis    RedisConfig
    App      AppConfig
}

type ServerConfig struct {
    Port string
    Host string
    Env  string
}

type DatabaseConfig struct {
    Host            string
    Port            string
    User            string
    Password        string
    Name            string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type JWTConfig struct {
    SecretKey        string
    ExpirationHours  int
}

type CORSConfig struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
}

type LoggerConfig struct {
    Level  string
    Format string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type AppConfig struct {
    PaginationDefaultLimit int
    PaginationMaxLimit    int
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, fmt.Errorf("error loading .env file: %v", err)
    }

    config := &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
            Host: getEnv("SERVER_HOST", "localhost"),
            Env:  getEnv("ENV", "development"),
        },
        Database: DatabaseConfig{
            Host:            getEnv("DB_HOST", "localhost"),
            Port:            getEnv("DB_PORT", "3306"),
            User:            getEnv("DB_USER", "root"),
            Password:        getEnv("DB_PASSWORD", ""),
            Name:            getEnv("DB_NAME", "university_db"),
            MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
            ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
        },
        JWT: JWTConfig{
            SecretKey:       getEnv("JWT_SECRET_KEY", ""),
            ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
        },
        CORS: CORSConfig{
            AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
            AllowedMethods: strings.Split(getEnv("ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
            AllowedHeaders: strings.Split(getEnv("ALLOWED_HEADERS", "Authorization,Content-Type"), ","),
        },
        Logger: LoggerConfig{
            Level:  getEnv("LOG_LEVEL", "debug"),
            Format: getEnv("LOG_FORMAT", "json"),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getEnvAsInt("REDIS_DB", 0),
        },
        App: AppConfig{
            PaginationDefaultLimit: getEnvAsInt("PAGINATION_DEFAULT_LIMIT", 10),
            PaginationMaxLimit:    getEnvAsInt("PAGINATION_MAX_LIMIT", 100),
        },
    }

    if err := validateConfig(config); err != nil {
        return nil, err
    }

    return config, nil
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value, exists := os.LookupEnv(key); exists {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func validateConfig(config *Config) error {
    if config.JWT.SecretKey == "" {
        return fmt.Errorf("JWT_SECRET_KEY is required")
    }

    if config.Database.Password == "" {
        return fmt.Errorf("DB_PASSWORD is required")
    }

    return nil
}

// Helper function to get Database DSN
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
        c.User,
        c.Password,
        c.Host,
        c.Port,
        c.Name,
    )
}

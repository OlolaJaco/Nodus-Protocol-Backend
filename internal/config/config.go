package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Email      EmailConfig
	Log        LogConfig
	CoreEngine CoreEngineConfig
	Stellar    StellarConfig
}

type AppConfig struct {
	Name        string `mapstructure:"APP_NAME"`
	Env         string `mapstructure:"APP_ENV"`
	FrontendURL string `mapstructure:"FRONTEND_URL"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"SERVER_PORT"`
	ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	PrivateKeyPath   string        `mapstructure:"JWT_PRIVATE_KEY_PATH"`
	PublicKeyPath    string        `mapstructure:"JWT_PUBLIC_KEY_PATH"`
	AccessTokenTTL   time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"JWT_REFRESH_TTL"`
	EmailTokenTTL    time.Duration `mapstructure:"JWT_EMAIL_TTL"`
	PasswordResetTTL time.Duration `mapstructure:"JWT_PASSWORD_RESET_TTL"`
}

type EmailConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	Username string `mapstructure:"SMTP_USERNAME"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"SMTP_FROM"`
	FromName string `mapstructure:"SMTP_FROM_NAME"`
}

type LogConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

// CoreEngineConfig holds connection settings for the Nodus Core Engine.
type CoreEngineConfig struct {
	URL     string        `mapstructure:"CORE_ENGINE_URL"`
	Timeout time.Duration `mapstructure:"CORE_ENGINE_TIMEOUT"`
}

// StellarConfig holds Stellar / Soroban connection settings.
// Used to proxy AMM pool reads from the Core Engine.
type StellarConfig struct {
	Network        string `mapstructure:"STELLAR_NETWORK"`
	HorizonURL     string `mapstructure:"HORIZON_URL"`
	SorobanRPCURL  string `mapstructure:"SOROBAN_RPC_URL"`
	PoolContractID string `mapstructure:"POOL_CONTRACT_ID"`
	PoolToken0     string `mapstructure:"POOL_TOKEN_0"`
	PoolToken1     string `mapstructure:"POOL_TOKEN_1"`
}

func (s StellarConfig) PoolConfigured() bool {
	return s.SorobanRPCURL != "" && s.PoolContractID != "" &&
		s.PoolToken0 != "" && s.PoolToken1 != ""
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("APP_NAME", "Nodus Protocol API")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("FRONTEND_URL", "http://localhost:3000")
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 120*time.Second)
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_PRIVATE_KEY_PATH", "./certs/private.pem")
	v.SetDefault("JWT_PUBLIC_KEY_PATH", "./certs/public.pem")
	v.SetDefault("JWT_ACCESS_TTL", 15*time.Minute)
	v.SetDefault("JWT_REFRESH_TTL", 7*24*time.Hour)
	v.SetDefault("JWT_EMAIL_TTL", 24*time.Hour)
	v.SetDefault("JWT_PASSWORD_RESET_TTL", 1*time.Hour)
	v.SetDefault("SMTP_PORT", 587)
	v.SetDefault("SMTP_FROM_NAME", "Nodus Protocol")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("CORE_ENGINE_URL", "http://localhost:8081")
	v.SetDefault("CORE_ENGINE_TIMEOUT", 15*time.Second)
	v.SetDefault("STELLAR_NETWORK", "testnet")
	v.SetDefault("HORIZON_URL", "https://horizon-testnet.stellar.org")

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		cfg.App = AppConfig{Name: v.GetString("APP_NAME"), Env: v.GetString("APP_ENV"), FrontendURL: v.GetString("FRONTEND_URL")}
		cfg.Server = ServerConfig{Port: v.GetString("SERVER_PORT"), ReadTimeout: v.GetDuration("SERVER_READ_TIMEOUT"), WriteTimeout: v.GetDuration("SERVER_WRITE_TIMEOUT"), IdleTimeout: v.GetDuration("SERVER_IDLE_TIMEOUT")}
		cfg.Database = DatabaseConfig{Host: v.GetString("DB_HOST"), Port: v.GetString("DB_PORT"), User: v.GetString("DB_USER"), Password: v.GetString("DB_PASSWORD"), Name: v.GetString("DB_NAME"), SSLMode: v.GetString("DB_SSLMODE")}
		cfg.Redis = RedisConfig{Host: v.GetString("REDIS_HOST"), Port: v.GetString("REDIS_PORT"), Password: v.GetString("REDIS_PASSWORD"), DB: v.GetInt("REDIS_DB")}
		cfg.JWT = JWTConfig{PrivateKeyPath: v.GetString("JWT_PRIVATE_KEY_PATH"), PublicKeyPath: v.GetString("JWT_PUBLIC_KEY_PATH"), AccessTokenTTL: v.GetDuration("JWT_ACCESS_TTL"), RefreshTokenTTL: v.GetDuration("JWT_REFRESH_TTL"), EmailTokenTTL: v.GetDuration("JWT_EMAIL_TTL"), PasswordResetTTL: v.GetDuration("JWT_PASSWORD_RESET_TTL")}
		cfg.Email = EmailConfig{Host: v.GetString("SMTP_HOST"), Port: v.GetInt("SMTP_PORT"), Username: v.GetString("SMTP_USERNAME"), Password: v.GetString("SMTP_PASSWORD"), From: v.GetString("SMTP_FROM"), FromName: v.GetString("SMTP_FROM_NAME")}
		cfg.Log = LogConfig{Level: v.GetString("LOG_LEVEL"), Format: v.GetString("LOG_FORMAT")}
		cfg.CoreEngine = CoreEngineConfig{URL: v.GetString("CORE_ENGINE_URL"), Timeout: v.GetDuration("CORE_ENGINE_TIMEOUT")}
		cfg.Stellar = StellarConfig{Network: v.GetString("STELLAR_NETWORK"), HorizonURL: v.GetString("HORIZON_URL"), SorobanRPCURL: v.GetString("SOROBAN_RPC_URL"), PoolContractID: v.GetString("POOL_CONTRACT_ID"), PoolToken0: v.GetString("POOL_TOKEN_0"), PoolToken1: v.GetString("POOL_TOKEN_1")}
	}

	return cfg, nil
}

func (c *Config) IsProd() bool {
	return c.App.Env == "production"
}

package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func LoadConfig() Config {
	return Config{
		Host:            getEnv("DATABASE_HOST", "localhost"),
		Port:            getEnv("DATABASE_PORT", "5432"),
		User:            getEnv("DATABASE_USER", "postgres"),
		Password:        getEnv("DATABASE_PASSWORD", ""),
		Name:            getEnv("DATABASE_NAME", "postgres"),
		SSLMode:         getEnv("DATABASE_SSL_MODE", "disable"),
		MaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 10),
		MaxOpenConns:    getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
		ConnMaxLifetime: time.Duration(getEnvAsInt("DATABASE_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
		ConnMaxIdleTime: time.Duration(getEnvAsInt("DATABASE_CONN_MAX_IDLE_MINUTES", 5)) * time.Minute,
	}
}

func Connect(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn(cfg)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	configurePool(sqlDB, cfg)
	return db, nil
}

func OpenSQL(cfg Config) (*sql.DB, error) {
	sqlDB, err := sql.Open("pgx", dsn(cfg))
	if err != nil {
		return nil, err
	}
	configurePool(sqlDB, cfg)
	return sqlDB, nil
}

func dsn(cfg Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
}

func configurePool(sqlDB *sql.DB, cfg Config) {
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func getEnvAsInt(key string, def int) int {
	val := getEnv(key, "")
	if val == "" {
		return def
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return parsed
}

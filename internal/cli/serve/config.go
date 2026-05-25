package serve

import (
	"fmt"
	"os"
)

type Config struct {
	MySQLDSN string // MYSQL_DSN
	HTTPAddr string // HTTP_ADDR, default ":8080"
	LogLevel string // LOG_LEVEL, default "info"
}

func LoadConfig() (Config, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return Config{}, fmt.Errorf("MYSQL_DSN is required")
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	lvl := os.Getenv("LOG_LEVEL")
	if lvl == "" {
		lvl = "info"
	}
	return Config{MySQLDSN: dsn, HTTPAddr: addr, LogLevel: lvl}, nil
}

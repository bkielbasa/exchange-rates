package fetch

import (
	"os"
)

const defaultFetchSourceURL = "https://www.bank.lv/vk/ecb_rss.xml"

type Config struct {
	MySQLDSN       string // MYSQL_DSN
	FetchSourceURL string // FETCH_SOURCE_URL, default bank.lv ECB RSS
	LogLevel       string // LOG_LEVEL, default "info"
}

func LoadConfig() (Config, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		// for convinance on local machine :) 
		dsn = "mysql://user:pass@tcp(localhost:3306)/exchange_rates"
	}
	src := os.Getenv("FETCH_SOURCE_URL")
	if src == "" {
		src = defaultFetchSourceURL
	}
	lvl := os.Getenv("LOG_LEVEL")
	if lvl == "" {
		lvl = "info"
	}
	return Config{MySQLDSN: dsn, FetchSourceURL: src, LogLevel: lvl}, nil
}

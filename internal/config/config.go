package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
	Tokens      map[string]string
}

func Load() (Config, error) {
	cfg := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: os.Getenv("HTTP_ADDR")}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "127.0.0.1:8080"
	}

	_, port, err := net.SplitHostPort(cfg.HTTPAddr)
	if err != nil {
		return Config{}, fmt.Errorf("HTTP_ADDR must be host:port")
	}

	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR must contain a port between 1 and 65535")
	}

	first, second := os.Getenv("DEV_TOKEN_TIM"), os.Getenv("DEV_TOKEN_HELEN")
	if first == "" || second == "" || first == second || strings.ContainsAny(first+second, " \t\r\n") {
		return Config{}, fmt.Errorf("DEV_TOKEN_TIM and DEV_TOKEN_HELEN must be distinct nonempty tokens without whitespace")
	}

	cfg.Tokens = map[string]string{first: "tim", second: "helen"}

	return cfg, nil
}

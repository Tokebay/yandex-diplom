package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress        string
	DatabaseURI       string
	AccrualSystemAddr string
}

func NewConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address and port to run the service")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "", "address of the accrual system")

	flag.Parse()

	cfg.parseEnv()

	return cfg
}

func (c *Config) parseEnv() {
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		c.RunAddress = envRunAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		c.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddr != "" {
		c.AccrualSystemAddr = envAccrualSystemAddr
	}
}

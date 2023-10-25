package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	RunAddress        string
	DatabaseURI       string
	AccrualSystemAddr string
	AccrualSystemPort string
}

func NewConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddress, "a", ":8080", "address and port to run the service")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "http://localhost:9090", "address of the accrual system")

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
	PrintUsage()
}

func PrintUsage() {
	fmt.Println("Usage: gofermart -a ADDRESS -d DATABASE_URI -r ACCRUAL_SYSTEM_ADDRESS")
	flag.PrintDefaults()
}

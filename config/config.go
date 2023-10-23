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
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address and port to run the service")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", "localhost:9090", "address of the accrual system")

	flag.Parse()

	cfg.parseEnv()

	return cfg
}

func (c *Config) parseEnv() {
	PrintUsage()

	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		c.RunAddress = envRunAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		c.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddr != "" {
		c.AccrualSystemPort = envAccrualSystemAddr
	}
}

func PrintUsage() {
	fmt.Println("Usage: gofermart -a ADDRESS -d DATABASE_URI -r ACCRUAL_SYSTEM_ADDRESS")
	flag.PrintDefaults()
}

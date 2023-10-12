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
}

func NewConfig() *Config {
	return &Config{
		RunAddress:        getEnv("RUN_ADDRESS", ":8080"),
		DatabaseURI:       getEnv("DATABASE_URI", "postgresql://username:password@localhost:5432/dbname"),
		AccrualSystemAddr: getEnv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8000"),
	}
}

func getEnv(key, fallbackValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallbackValue
}

func ParseFlags(cfg *Config) {
	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "address and port to run the service")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "database connection URI")
	flag.StringVar(&cfg.AccrualSystemAddr, "r", cfg.AccrualSystemAddr, "address of the accrual system")
	flag.Parse()
}

func PrintUsage() {
	fmt.Println("Usage: gofermart -a ADDRESS -d DATABASE_URI -r ACCRUAL_SYSTEM_ADDRESS")
	flag.PrintDefaults()
}

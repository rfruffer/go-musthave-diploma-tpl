package config

import (
	"flag"
	"os"
)

type Config struct {
	StartHost string
	DBDSN     string
	SecretKey string
	Accrual   string
}

func ParseFlags() *Config {
	startHost := flag.String("a", "0.0.0.0:8080", "address and port to run server")
	accrual := flag.String("r", "0.0.0.0:8080", "address to run accrual")
	dbDSN := flag.String("d", "", "database DSN for PostgreSQL")
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "verysecretkey"
	}

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		*startHost = envRunAddr
	}
	if envAccrual := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrual != "" {
		*accrual = envAccrual
	}
	if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
		*dbDSN = envDB
	}

	return &Config{
		StartHost: *startHost,
		DBDSN:     *dbDSN,
		Accrual:   *accrual,
		SecretKey: secretKey,
	}
}

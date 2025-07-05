package config

import (
	"flag"
	"os"
)

type Config struct {
	StartHost  string
	ResultHost string
	FilePath   string
	DBDSN      string
	Storage    string
	SecretKey  string
}

func ParseFlags() *Config {
	startHost := flag.String("a", "0.0.0.0:8080", "address and port to run server")
	resultHost := flag.String("b", "http://localhost:8080", "base URL for shortened links")
	filePath := flag.String("f", "", "path to file storage")
	dbDSN := flag.String("d", "", "database DSN for PostgreSQL")
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "verysecretkey"
	}

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		*startHost = envRunAddr
	}
	if envResultHost := os.Getenv("BASE_URL"); envResultHost != "" {
		*resultHost = envResultHost
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		*filePath = envFilePath
	}
	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		*dbDSN = envDB
	}

	storage := ""
	if *dbDSN != "" {
		storage = "postgres"
	}

	return &Config{
		StartHost:  *startHost,
		ResultHost: *resultHost,
		FilePath:   *filePath,
		DBDSN:      *dbDSN,
		Storage:    storage,
		SecretKey:  secretKey,
	}
}

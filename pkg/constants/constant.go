package constants

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type EnvVars struct {
	DB_URI        string
	SSL_CERT_PATH string
	SALT          string
	JWT_SECRET    string
	SMTP_HOST     string
	SMTP_PORT     string
	SMTP_USER     string
	SMTP_PASS     string
}

func LoadEnv() (*EnvVars, error) {
	envPath, err := filepath.Abs(".env")
	if err == nil {
		if loadErr := godotenv.Load(envPath); loadErr != nil {
			log.Printf("No .env file loaded from %s (this is okay on Render): %v", envPath, loadErr)
		} else {
			log.Println(".env file loaded successfully")
		}
	} else {
		log.Printf("Couldn't resolve path to .env (this is okay on Render): %v", err)
	}

	db_uri := os.Getenv("DB_URI")
	ssl_cert_path := os.Getenv("SSL_CERT_PATH")
	if ssl_cert_path != "" {
		ssl_cert_path, err = filepath.Abs(ssl_cert_path)
		if err != nil {
			return nil, fmt.Errorf("error resolving SSL_CERT_PATH: %v", err)
		}
		log.Printf("Resolved SSL cert path: %s", ssl_cert_path)
	}

	salt := os.Getenv("SALT")
	jwt_secret := os.Getenv("JWT_SECRET")
	smtp_host := os.Getenv("SMTP_HOST")
	smtp_port := os.Getenv("SMTP_PORT")
	smtp_user := os.Getenv("SMTP_USER")
	smtp_pass := os.Getenv("SMTP_PASS")

	log.Printf("SMTP_HOST from env: '%s'", smtp_host)
	log.Printf("SMTP_PORT from env: '%s'", smtp_port)
	log.Printf("SMTP_USER from env: '%s'", smtp_user)
	log.Printf("SMTP_PASS from env: '%s'", smtp_pass)

	missing := func(name, val string) error {
		if val == "" {
			return fmt.Errorf("%s not set in the environment", name)
		}
		return nil
	}
	for _, err := range []error{
		missing("DB_URI", db_uri),
		missing("SSL_CERT_PATH", ssl_cert_path),
		missing("SALT", salt),
		missing("JWT_SECRET", jwt_secret),
		missing("SMTP_HOST", smtp_host),
		missing("SMTP_PORT", smtp_port),
		missing("SMTP_USER", smtp_user),
		missing("SMTP_PASS", smtp_pass),
	} {
		if err != nil {
			return nil, err
		}
	}

	return &EnvVars{
		DB_URI:        db_uri,
		SSL_CERT_PATH: ssl_cert_path,
		SALT:          salt,
		JWT_SECRET:    jwt_secret,
		SMTP_HOST:     smtp_host,
		SMTP_PORT:     smtp_port,
		SMTP_USER:     smtp_user,
		SMTP_PASS:     smtp_pass,
	}, nil
}

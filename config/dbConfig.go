package config

import (
	"fmt"
	"log"
	"path/filepath"
	"realTimeEditor/internal/model"
	"realTimeEditor/pkg/constants"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	envVars, err := constants.LoadEnv()
	if err != nil {
		log.Printf("Error loading environment variables: %v\n", err)
		return
	}

	sslCertPath, err := filepath.Abs(envVars.SSL_CERT_PATH)
	if err != nil {
		log.Printf("Error resolving SSL_CERT_PATH: %v\n", err)
		return
	}

	dbUri := fmt.Sprintf("%s&sslrootcert=%s", envVars.DB_URI, sslCertPath)
	db, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: %v", err))
	}

	if err := db.AutoMigrate(
		&model.User{}, &model.Document{}, &model.DocumentAccess{}, &model.ForgotPassword{},
	); err != nil {
		panic(fmt.Sprintf("Error during migration: %v", err))
	}

	DB = db
	fmt.Println("Database connection initialized successfully")
}

func GetDB() *gorm.DB {
	return DB
}

func CloseDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying DB: %w", err)
	}
	return sqlDB.Close()
}

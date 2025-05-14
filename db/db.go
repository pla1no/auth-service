package db

import (
	"auth-service/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SLLMode  string
}

var DB *gorm.DB

func NewPostgresDB(cfg Config) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=%s", cfg.Username, cfg.Password, cfg.DBName, cfg.Port, cfg.SLLMode),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %e", err))
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(fmt.Sprintf("Failed to automigrate for database: %e", err))
	}

	fmt.Println("Migrated database")

	DB = db
}

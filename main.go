package main

import (
	"auth-service/db"
	"auth-service/routes"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	if err := InitConfig(); err != nil {
		log.Fatalf("error initializing config %s", err.Error())
	}

	db.NewPostgresDB(db.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.dbname"),
		SLLMode:  viper.GetString("db.sslmode"),
	})

	r := gin.Default()

	routes.AuthRoutes(r)

	r.Run(":8080")
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

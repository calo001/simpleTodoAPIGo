package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"os"
	"todoAPI/config"
	"todoAPI/migration"
	"todoAPI/route"
)

func init() {
	db := config.Init()
	migration.Migrate(db)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := route.SetupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}
package main

import (
	"account-microservice/config"
	"account-microservice/models"
	"account-microservice/routes"

	"github.com/gin-gonic/gin"
)

func main() {
    config.InitDB()
    config.DB.AutoMigrate(
        &models.Account{},
        &models.Role{},
        &models.Token{},
        &models.Doctor{},
        &models.Specialization{},
    )
    gin.SetMode(gin.DebugMode) 
    r := gin.Default()
    r.Use(gin.Logger())
    routes.InitAuthRoutes(r)
    routes.InitAccountRoutes(r)
    routes.InitDoctorRoutes(r)

    r.Run(":8080")
}

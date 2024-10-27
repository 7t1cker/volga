package main

import (
	"hospital_service/config"
	"hospital_service/models"
	"hospital_service/routes"
	"hospital_service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
    config.InitDB()
    config.DB.AutoMigrate(&models.Hospital{}, &models.Room{})

    accountService := utils.NewAccountService()

    r := gin.Default()

    routes.InitHospitalRoutes(r, accountService)

    r.Run(":8081")
}

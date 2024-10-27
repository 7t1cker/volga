package main

import (
	"timetable_service/config"
	"timetable_service/models"
	"timetable_service/routes"
	"timetable_service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
    config.InitDB()
    config.DB.AutoMigrate(&models.Timetable{}, &models.Appointment{})

    accountService := utils.NewAccountService()

    r := gin.Default()

    routes.InitTimetableRoutes(r, accountService)

    r.Run(":8082") 
}

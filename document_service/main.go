package main

import (
	"document_service/config"
	"document_service/models"
	"document_service/routes"
	"document_service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()
	config.DB.AutoMigrate(&models.History{})
	accountService := utils.NewAccountService()
	r := gin.Default()
	routes.InitHistoryRoutes(r, accountService)
	r.Run(":8083")
}

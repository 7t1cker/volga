package routes

import (
	"hospital_service/controllers"
	"hospital_service/middlewares"
	"hospital_service/utils"

	"github.com/gin-gonic/gin"
)

func InitHospitalRoutes(r *gin.Engine, accountService *utils.AccountService) {
    hospitalRoutes := r.Group("/api/Hospitals")
    {
        hospitalRoutes.GET("/", middlewares.AuthMiddleware(accountService), controllers.GetHospitals)
        hospitalRoutes.GET("/:id", middlewares.AuthMiddleware(accountService), controllers.GetHospitalByID)
        hospitalRoutes.GET("/:id/Rooms", middlewares.AuthMiddleware(accountService), controllers.GetHospitalRooms)

        hospitalRoutes.POST("/", middlewares.AuthMiddleware(accountService), middlewares.AdminMiddleware(accountService), controllers.CreateHospital)
        hospitalRoutes.PUT("/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminMiddleware(accountService), controllers.UpdateHospital)
        hospitalRoutes.DELETE("/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminMiddleware(accountService), controllers.DeleteHospital)
    }
}

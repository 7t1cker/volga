package routes

import (
	"account-microservice/controllers"
	"account-microservice/middlewares"

	"github.com/gin-gonic/gin"
)

func InitDoctorRoutes(r *gin.Engine) {
    doctorRoutes := r.Group("/api/Doctors")
    {
        doctorRoutes.GET("/", middlewares.JWTAuthMiddleware(), controllers.GetDoctors)
        doctorRoutes.GET("/:id", middlewares.JWTAuthMiddleware(), controllers.GetDoctorByID)
        doctorRoutes.POST("/", middlewares.JWTAuthMiddleware(), middlewares.AdminMiddleware(), controllers.CreateDoctor)
    }
}

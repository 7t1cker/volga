package routes

import (
	"account-microservice/controllers"
	"account-microservice/middlewares"

	"github.com/gin-gonic/gin"
)

func InitAuthRoutes(r *gin.Engine) {
    authRoutes := r.Group("/api/Authentication")
    {
        authRoutes.POST("/SignUp", controllers.SignUp)
        authRoutes.POST("/SignIn", controllers.SignIn)
        authRoutes.PUT("/SignOut", middlewares.JWTAuthMiddleware(), controllers.SignOut)
        authRoutes.GET("/Validate", controllers.ValidateToken)
        authRoutes.POST("/Refresh", controllers.RefreshToken)
    }
}

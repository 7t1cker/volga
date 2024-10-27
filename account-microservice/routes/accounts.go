package routes

import (
	"account-microservice/controllers"
	"account-microservice/middlewares"

	"github.com/gin-gonic/gin"
)

func InitAccountRoutes(r *gin.Engine) {
    accountRoutes := r.Group("/api/Accounts")
    {
        accountRoutes.GET("/Me", middlewares.JWTAuthMiddleware(), controllers.GetCurrentAccount)
        accountRoutes.PUT("/Update", middlewares.JWTAuthMiddleware(), controllers.UpdateCurrentAccount)
        accountRoutes.GET("/", middlewares.JWTAuthMiddleware(), middlewares.AdminMiddleware(), controllers.GetAllAccounts)
        accountRoutes.POST("/", middlewares.JWTAuthMiddleware(), middlewares.AdminMiddleware(), controllers.CreateAccount)
        accountRoutes.PUT("/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminMiddleware(), controllers.UpdateAccount)
        accountRoutes.DELETE("/:id", middlewares.JWTAuthMiddleware(), middlewares.AdminMiddleware(), controllers.DeleteAccount)
        accountRoutes.GET("/:id/roles", middlewares.JWTAuthMiddleware(), controllers.CheckUserRole)

    }
}

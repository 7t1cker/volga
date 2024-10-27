package routes

import (
	"document_service/controllers"
	"document_service/middlewares"
	"document_service/utils"

	"github.com/gin-gonic/gin"
)

func InitHistoryRoutes(r *gin.Engine, accountService *utils.AccountService) {
    historyController := &controllers.HistoryController{
        AccountService: accountService,
    }

    historyRoutes := r.Group("/api/History")
    {
        historyRoutes.GET("/Account/:id",middlewares.AuthMiddleware(accountService),historyController.GetHistoryByAccountID,)
        historyRoutes.GET("/:id", middlewares.AuthMiddleware(accountService),historyController.GetHistoryByID,)
        historyRoutes.POST("", middlewares.AuthMiddleware(accountService),middlewares.RoleMiddleware(accountService, []string{"admin", "manager", "doctor"}),historyController.CreateHistory,)
        historyRoutes.PUT("/:id", middlewares.AuthMiddleware(accountService),middlewares.RoleMiddleware(accountService, []string{"admin", "manager", "doctor"}), historyController.UpdateHistory, )
    }
}

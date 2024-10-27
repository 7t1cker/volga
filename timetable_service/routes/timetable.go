package routes

import (
	"timetable_service/controllers"
	"timetable_service/middlewares"
	"timetable_service/utils"

	"github.com/gin-gonic/gin"
)

func InitTimetableRoutes(r *gin.Engine, accountService *utils.AccountService) {
    timetableRoutes := r.Group("/api/Timetable")
    {
        timetableRoutes.POST("/", middlewares.AuthMiddleware(accountService), middlewares.AdminOrManagerMiddleware(accountService), controllers.CreateTimetable)
        timetableRoutes.PUT("/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminOrManagerMiddleware(accountService), controllers.UpdateTimetable)
        timetableRoutes.DELETE("/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminOrManagerMiddleware(accountService), controllers.DeleteTimetable)
        timetableRoutes.DELETE("/Doctor/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminOrManagerMiddleware(accountService), controllers.DeleteTimetableByDoctor)
        timetableRoutes.DELETE("/Hospital/:id", middlewares.AuthMiddleware(accountService), middlewares.AdminOrManagerMiddleware(accountService), controllers.DeleteTimetableByHospital)

        timetableRoutes.GET("/Hospital/:id", middlewares.AuthMiddleware(accountService), controllers.GetTimetableByHospital)
        timetableRoutes.GET("/Doctor/:id", middlewares.AuthMiddleware(accountService), controllers.GetTimetableByDoctor)
        timetableRoutes.GET("/Hospital/:id/Room/:room", middlewares.AuthMiddleware(accountService), middlewares.AdminManagerOrDoctorMiddleware(accountService), controllers.GetTimetableByRoom)

        timetableRoutes.GET("/:id/Appointments", middlewares.AuthMiddleware(accountService), controllers.GetAvailableAppointments)
        timetableRoutes.POST("/:id/Appointments", middlewares.AuthMiddleware(accountService), controllers.CreateAppointment)
    }

    appointmentRoutes := r.Group("/api/Appointment")
    {
        appointmentRoutes.DELETE("/:id", middlewares.AuthMiddleware(accountService), controllers.DeleteAppointment)
    }
}

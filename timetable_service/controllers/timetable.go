package controllers

import (
	"net/http"
	"time"

	"timetable_service/config"
	"timetable_service/models"
	"timetable_service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var thirtyMinutes = time.Minute * 30
var twelveHours = time.Hour * 12

func CreateTimetable(c *gin.Context) {
	var input struct {
		HospitalID uint      `json:"hospitalId" binding:"required"`
		DoctorID   uint      `json:"doctorId" binding:"required"`
		From       time.Time `json:"from" binding:"required"`
		To         time.Time `json:"to" binding:"required"`
		Room       string    `json:"room" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
		return
	}
	if len(token) <= len("Bearer ") || token[:len("Bearer ")] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization token format"})
		return
	}
	token = token[len("Bearer "):]

	if !validateTime(input.From) || !validateTime(input.To) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time must be in 30-minute increments and seconds must be zero"})
		return
	}

	if input.To.Before(input.From) || input.To.Equal(input.From) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'to' must be greater than 'from'"})
		return
	}

	if input.To.Sub(input.From) > twelveHours {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time difference between 'from' and 'to' must not exceed 12 hours"})
		return
	}

	hospitalService := utils.NewHospitalService()
	isValid, err := hospitalService.ValidateHospitalAndRoom(input.HospitalID, input.Room, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital or room"})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start database transaction"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	var existingRoomTimetable models.Timetable
	roomConflict := tx.Where(`hospital_id = ? AND room = ? AND ("from" < ? AND "to" > ?)`, input.HospitalID, input.Room, input.To, input.From).First(&existingRoomTimetable).Error
	if roomConflict == nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is already booked for this time period"})
		return
	} else if roomConflict != gorm.ErrRecordNotFound {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check room availability"})
		return
	}

	var existingDoctorTimetable models.Timetable
	doctorConflict := tx.Where(`doctor_id = ? AND ("from" < ? AND "to" > ?)`, input.DoctorID, input.To, input.From).First(&existingDoctorTimetable).Error
	if doctorConflict == nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor is already booked for this time period"})
		return
	} else if doctorConflict != gorm.ErrRecordNotFound {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check doctor availability"})
		return
	}

	doctorService := utils.NewDoctorService()
	isDoctor, err := doctorService.ValidateDoctor(input.DoctorID, token)
	if err != nil || !isDoctor {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Specified user is not a doctor"})
		return
	}

	timetable := models.Timetable{
		HospitalID: input.HospitalID,
		DoctorID:   input.DoctorID,
		From:       input.From,
		To:         input.To,
		Room:       input.Room,
	}

	if err := tx.Create(&timetable).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create timetable"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.Status(http.StatusCreated)
}

func validateTime(t time.Time) bool {
	return t.Minute()%30 == 0 && t.Second() == 0 && t.Nanosecond() == 0
}

func UpdateTimetable(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		HospitalID uint      `json:"hospitalId"`
		DoctorID   uint      `json:"doctorId"`
		From       time.Time `json:"from"`
		To         time.Time `json:"to"`
		Room       string    `json:"room"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
		return
	}
	if len(token) <= len("Bearer ") || token[:len("Bearer ")] != "Bearer " {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization token format"})
		return
	}
	token = token[len("Bearer "):]

	if !input.From.IsZero() && !validateTime(input.From) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time must be in 30-minute increments and seconds must be zero"})
		return
	}

	if !input.To.IsZero() && !validateTime(input.To) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time must be in 30-minute increments and seconds must be zero"})
		return
	}

	if !input.From.IsZero() && !input.To.IsZero() {
		if input.To.Before(input.From) || input.To.Equal(input.From) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "'to' must be greater than 'from'"})
			return
		}

		if input.To.Sub(input.From) > twelveHours {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Time difference between 'from' and 'to' must not exceed 12 hours"})
			return
		}
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start database transaction"})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	var timetable models.Timetable
	if err := tx.Preload("Appointments").First(&timetable, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Timetable not found"})
		return
	}

	if len(timetable.Appointments) > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update timetable with existing appointments"})
		return
	}

	if input.HospitalID != 0 || input.Room != "" {
		hospitalService := utils.NewHospitalService()
		hospitalID := input.HospitalID
		room := input.Room

		if hospitalID == 0 {
			hospitalID = timetable.HospitalID
		}
		if room == "" {
			room = timetable.Room
		}

		isValid, err := hospitalService.ValidateHospitalAndRoom(hospitalID, room, token)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !isValid {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hospital or room"})
			return
		}
	}

	if input.DoctorID != 0 {
		doctorService := utils.NewDoctorService()
		isDoctor, err := doctorService.ValidateDoctor(input.DoctorID, token)
		if err != nil || !isDoctor {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Specified user is not a doctor"})
			return
		}
	}

	if (!input.From.IsZero() || !input.To.IsZero() || input.HospitalID != 0 || input.Room != "") {
		newHospitalID := timetable.HospitalID
		newRoom := timetable.Room
		newFrom := timetable.From
		newTo := timetable.To

		if input.HospitalID != 0 {
			newHospitalID = input.HospitalID
		}
		if input.Room != "" {
			newRoom = input.Room
		}
		if !input.From.IsZero() {
			newFrom = input.From
		}
		if !input.To.IsZero() {
			newTo = input.To
		}

		var existingRoomTimetable models.Timetable
		roomConflict := tx.Where(
			`hospital_id = ? AND room = ? AND "from" < ? AND "to" > ? AND id != ?`,
			newHospitalID,
			newRoom,
			newTo,
			newFrom,
			id,
		).First(&existingRoomTimetable).Error

		if roomConflict == nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Room is already booked for this time period"})
			return
		} else if roomConflict != gorm.ErrRecordNotFound {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check room availability"})
			return
		}
	}

	if (!input.From.IsZero() || !input.To.IsZero() || input.DoctorID != 0) {
		newDoctorID := timetable.DoctorID
		newFrom := timetable.From
		newTo := timetable.To

		if input.DoctorID != 0 {
			newDoctorID = input.DoctorID
		}
		if !input.From.IsZero() {
			newFrom = input.From
		}
		if !input.To.IsZero() {
			newTo = input.To
		}

		var existingDoctorTimetable models.Timetable
		doctorConflict := tx.Where(
			`doctor_id = ? AND "from" < ? AND "to" > ? AND id != ?`,
			newDoctorID,
			newTo,
			newFrom,
			id,
		).First(&existingDoctorTimetable).Error

		if doctorConflict == nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor is already booked for this time period"})
			return
		} else if doctorConflict != gorm.ErrRecordNotFound {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check doctor availability"})
			return
		}
	}

	if input.HospitalID != 0 {
		timetable.HospitalID = input.HospitalID
	}
	if input.DoctorID != 0 {
		timetable.DoctorID = input.DoctorID
	}
	if !input.From.IsZero() {
		timetable.From = input.From
	}
	if !input.To.IsZero() {
		timetable.To = input.To
	}
	if input.Room != "" {
		timetable.Room = input.Room
	}

	if err := tx.Save(&timetable).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update timetable"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteTimetable(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.Timetable{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete timetable"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteTimetableByDoctor(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("doctor_id = ?", id).Delete(&models.Timetable{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete timetables"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteTimetableByHospital(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("hospital_id = ?", id).Delete(&models.Timetable{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete timetables"})
		return
	}

	c.Status(http.StatusOK)
}

func GetTimetableByHospital(c *gin.Context) {
	id := c.Param("id")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var timetables []models.Timetable
	query := config.DB.Where("hospital_id = ?", id)

	if fromStr != "" && toStr != "" {
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' parameter"})
			return
		}
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' parameter"})
			return
		}
		query = query.Where("`from` >= ? AND `to` <= ?", fromTime, toTime)
	}

	if err := query.Find(&timetables).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timetables"})
		return
	}

	c.JSON(http.StatusOK, timetables)
}

func GetTimetableByDoctor(c *gin.Context) {
	id := c.Param("id")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var timetables []models.Timetable
	query := config.DB.Where("doctor_id = ?", id)

	if fromStr != "" && toStr != "" {
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' parameter"})
			return
		}
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' parameter"})
			return
		}
		query = query.Where("`from` >= ? AND `to` <= ?", fromTime, toTime)
	}

	if err := query.Find(&timetables).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timetables"})
		return
	}

	c.JSON(http.StatusOK, timetables)
}

func GetTimetableByRoom(c *gin.Context) {
	hospitalID := c.Param("id")
	room := c.Param("room")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var timetables []models.Timetable
	query := config.DB.Where("hospital_id = ? AND room = ?", hospitalID, room)

	if fromStr != "" && toStr != "" {
		fromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' parameter"})
			return
		}
		toTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' parameter"})
			return
		}
		query = query.Where("`from` >= ? AND `to` <= ?", fromTime, toTime)
	}

	if err := query.Find(&timetables).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timetables"})
		return
	}

	c.JSON(http.StatusOK, timetables)
}

func GetAvailableAppointments(c *gin.Context) {
	id := c.Param("id")

	var timetable models.Timetable
	if err := config.DB.Preload("Appointments").First(&timetable, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Timetable not found"})
		return
	}

	slots := generateTimeSlots(timetable.From, timetable.To)
	bookedSlots := make(map[time.Time]bool)
	for _, appointment := range timetable.Appointments {
		bookedSlots[appointment.Time] = true
	}

	var availableSlots []time.Time
	for _, slot := range slots {
		if !bookedSlots[slot] {
			availableSlots = append(availableSlots, slot)
		}
	}

	c.JSON(http.StatusOK, availableSlots)
}

func generateTimeSlots(from time.Time, to time.Time) []time.Time {
	var slots []time.Time
	for t := from; t.Before(to); t = t.Add(thirtyMinutes) {
		slots = append(slots, t)
	}
	return slots
}

func CreateAppointment(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Time time.Time `json:"time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, _ := c.Get("accessToken")
	accountService := utils.NewAccountService()
	userID, err := accountService.GetUserID(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user ID", "details": err.Error()})
		return
	}

	var timetable models.Timetable
	if err := config.DB.Preload("Appointments").First(&timetable, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Timetable not found"})
		return
	}

	if input.Time.Before(timetable.From) || input.Time.After(timetable.To) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Selected time is outside of timetable range"})
		return
	}

	if !validateTime(input.Time) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time must be in 30-minute increments and seconds must be zero"})
		return
	}

	var existingAppointment models.Appointment
	if err := config.DB.Where("timetable_id = ? AND time = ?", id, input.Time).First(&existingAppointment).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Time slot already booked"})
		return
	}

	appointment := models.Appointment{
		TimetableID: timetable.ID,
		UserID:      userID,
		Time:        input.Time,
	}

	if err := config.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
		return
	}

	c.Status(http.StatusCreated)
}

func DeleteAppointment(c *gin.Context) {
	id := c.Param("id")
	accessToken, _ := c.Get("accessToken")
	accountService := utils.NewAccountService()
	userID, err := accountService.GetUserID(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user ID", "details": err.Error()})
		return
	}

	var appointment models.Appointment
	if err := config.DB.First(&appointment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	roles, err := accountService.GetUserRoles(accessToken.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user roles", "details": err.Error()})
		return
	}

	isAdminOrManager := false
	for _, role := range roles {
		if role == "admin" || role == "manager" {
			isAdminOrManager = true
			break
		}
	}

	if appointment.UserID != userID && !isAdminOrManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to cancel this appointment"})
		return
	}

	if err := config.DB.Delete(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete appointment"})
		return
	}

	c.Status(http.StatusOK)
}

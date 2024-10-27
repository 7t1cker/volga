package controllers

import (
	"net/http"
	"strconv"

	"hospital_service/config"
	"hospital_service/models"

	"github.com/gin-gonic/gin"
)

func GetHospitals(c *gin.Context) {
	fromStr := c.Query("from")
	countStr := c.Query("count")

	var from, count int
	var err error

	if fromStr != "" {
		from, err = strconv.Atoi(fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' parameter"})
			return
		}
	}

	if countStr != "" {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'count' parameter"})
			return
		}
	}

	var hospitals []models.Hospital
	query := config.DB.Preload("Rooms")

	if count > 0 {
		query = query.Offset(from).Limit(count)
	}

	if err := query.Find(&hospitals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve hospitals"})
		return
	}

	c.JSON(http.StatusOK, hospitals)
}

func GetHospitalByID(c *gin.Context) {
	id := c.Param("id")

	var hospital models.Hospital
	if err := config.DB.Preload("Rooms").First(&hospital, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hospital not found"})
		return
	}

	c.JSON(http.StatusOK, hospital)
}

func GetHospitalRooms(c *gin.Context) {
	id := c.Param("id")

	var rooms []models.Room
	if err := config.DB.Where("hospital_id = ?", id).Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rooms"})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func CreateHospital(c *gin.Context) {
	var input struct {
		Name         string   `json:"name" binding:"required"`
		Address      string   `json:"address"`
		ContactPhone string   `json:"contactPhone"`
		Rooms        []string `json:"rooms"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var rooms []models.Room
	for _, roomName := range input.Rooms {
		rooms = append(rooms, models.Room{Name: roomName})
	}

	hospital := models.Hospital{
		Name:         input.Name,
		Address:      input.Address,
		ContactPhone: input.ContactPhone,
		Rooms:        rooms,
	}

	if err := config.DB.Create(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hospital"})
		return
	}

	c.Status(http.StatusCreated)
}

func UpdateHospital(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Name         string   `json:"name"`
		Address      string   `json:"address"`
		ContactPhone string   `json:"contactPhone"`
		Rooms        []string `json:"rooms"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hospital models.Hospital
	if err := config.DB.Preload("Rooms").First(&hospital, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hospital not found"})
		return
	}

	if input.Name != "" {
		hospital.Name = input.Name
	}
	if input.Address != "" {
		hospital.Address = input.Address
	}
	if input.ContactPhone != "" {
		hospital.ContactPhone = input.ContactPhone
	}
	if input.Rooms != nil {
		if err := config.DB.Where("hospital_id = ?", hospital.ID).Delete(&models.Room{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rooms"})
			return
		}
		var rooms []models.Room
		for _, roomName := range input.Rooms {
			rooms = append(rooms, models.Room{Name: roomName, HospitalID: hospital.ID})
		}
		hospital.Rooms = rooms
	}

	if err := config.DB.Save(&hospital).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update hospital"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteHospital(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("id = ?", id).Delete(&models.Hospital{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hospital"})
		return
	}

	c.Status(http.StatusOK)
}

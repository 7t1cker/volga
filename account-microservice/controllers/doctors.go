package controllers

import (
	"net/http"
	"strconv"

	"account-microservice/config"
	"account-microservice/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetDoctors(c *gin.Context) {
	nameFilter := c.Query("nameFilter")
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

	var doctors []models.Account
	query := config.DB.Preload("Specializations").Preload("Roles")

	query = query.Joins("JOIN account_roles ON accounts.id = account_roles.account_id").
		Joins("JOIN roles ON roles.id = account_roles.role_id").
		Where("roles.name = ?", "doctor")

	if nameFilter != "" {
		query = query.Where("accounts.first_name ILIKE ? OR accounts.last_name ILIKE ?", "%"+nameFilter+"%", "%"+nameFilter+"%")
	}

	if count > 0 {
		query = query.Offset(from).Limit(count)
	}

	if err := query.Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve doctors"})
		return
	}

	c.JSON(http.StatusOK, doctors)
}

func GetDoctorByID(c *gin.Context) {
	id := c.Param("id")

	var doctor models.Account
	if err := config.DB.Preload("Specializations").Preload("Roles").First(&doctor, id).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor not found"})
		return
	}

	isDoctor := false
	for _, role := range doctor.Roles {
		if role.Name == "doctor" {
			isDoctor = true
			break
		}
	}

	if !isDoctor {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account is not a doctor"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             doctor.ID,
		"lastName":       doctor.LastName,
		"firstName":      doctor.FirstName,
		"specializations": doctor.Specializations,
	})
}

func CreateDoctor(c *gin.Context) {
	rolesInterface, exists := c.Get("roles")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Roles not found"})
		return
	}

	roles := rolesInterface.([]interface{})
	isAdmin := false
	for _, role := range roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin privileges required"})
		return
	}

	var input struct {
		LastName       string   `json:"lastName" binding:"required"`
		FirstName      string   `json:"firstName" binding:"required"`
		Username       string   `json:"username" binding:"required"`
		Password       string   `json:"password" binding:"required"`
		Specializations []string `json:"specializations" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password encryption failed"})
		return
	}

	var doctorRole models.Role
	if err := config.DB.FirstOrCreate(&doctorRole, models.Role{Name: "doctor"}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve or create doctor role"})
		return
	}

	var specializations []*models.Specialization
	for _, specName := range input.Specializations {
		var specialization models.Specialization
		if err := config.DB.FirstOrCreate(&specialization, models.Specialization{Name: specName}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create specialization"})
			return
		}
		specializations = append(specializations, &specialization)
	}

	doctor := models.Account{
		LastName:       input.LastName,
		FirstName:      input.FirstName,
		Username:       input.Username,
		Password:       string(passwordHash),
		Roles:          []*models.Role{&doctorRole},
		Specializations: specializations,
	}

	if err := config.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.Status(http.StatusCreated)
}

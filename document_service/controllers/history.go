package controllers

import (
	"document_service/config"
	"document_service/models"
	"document_service/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type HistoryController struct {
	AccountService *utils.AccountService
}

func (h *HistoryController) GetHistoryByAccountID(c *gin.Context) {
	idParam := c.Param("id")
	accountID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	accessToken := c.GetString("accessToken")
	currentUserID, err := h.AccountService.GetAccountID(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get current user ID"})
		return
	}

	roles, err := h.AccountService.GetRolesByID(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user roles"})
		return
	}

	isOwner := uint(accountID) == currentUserID
	isDoctor := containsRole(roles, "doctor")

	if !isOwner && !isDoctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var histories []models.History
	if err := config.DB.Where("pacient_id = ?", accountID).Find(&histories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve histories"})
		return
	}

	c.JSON(http.StatusOK, histories)
}

func (h *HistoryController) GetHistoryByID(c *gin.Context) {
	idParam := c.Param("id")
	historyID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID"})
		return
	}

	var history models.History
	if err := config.DB.First(&history, historyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
		return
	}

	accessToken := c.GetString("accessToken")
	currentUserID, err := h.AccountService.GetAccountID(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get current user ID"})
		return
	}

	roles, err := h.AccountService.GetRolesByID(accessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user roles"})
		return
	}

	isOwner := history.PacientID == currentUserID
	isDoctor := containsRole(roles, "doctor")

	if !isOwner && !isDoctor {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *HistoryController) CreateHistory(c *gin.Context) {
	var input struct {
		Date       string `json:"date" binding:"required"`
		PacientID  uint   `json:"pacientId" binding:"required"`
		HospitalID uint   `json:"hospitalId" binding:"required"`
		DoctorID   uint   `json:"doctorId" binding:"required"`
		Room       string `json:"room" binding:"required"`
		Data       string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	accessToken := c.GetString("accessToken")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token not found"})
		return
	}

	_, err = h.AccountService.GetRolesByAccountID(input.PacientID, accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get pacient roles"})
		return
	}

	history := models.History{
		Date:       date,
		PacientID:  input.PacientID,
		HospitalID: input.HospitalID,
		DoctorID:   input.DoctorID,
		Room:       input.Room,
		Data:       input.Data,
	}

	if err := config.DB.Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create history"})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *HistoryController) UpdateHistory(c *gin.Context) {
	idParam := c.Param("id")
	historyID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID"})
		return
	}

	var history models.History
	if err := config.DB.First(&history, historyID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
		return
	}

	var input struct {
		Date       string `json:"date"`
		PacientID  uint   `json:"pacientId"`
		HospitalID uint   `json:"hospitalId"`
		DoctorID   uint   `json:"doctorId"`
		Room       string `json:"room"`
		Data       string `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Date != "" {
		date, err := time.Parse(time.RFC3339, input.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		history.Date = date
	}
	if input.PacientID != 0 {
		accessToken := c.GetString("accessToken")
		pacientRoles, err := h.AccountService.GetRolesByAccountID(input.PacientID, accessToken)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get pacient roles"})
			return
		}

		if !containsRole(pacientRoles, "user") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Pacient must have role 'user'"})
			return
		}
		history.PacientID = input.PacientID
	}
	if input.HospitalID != 0 {
		history.HospitalID = input.HospitalID
	}
	if input.DoctorID != 0 {
		history.DoctorID = input.DoctorID
	}
	if input.Room != "" {
		history.Room = input.Room
	}
	if input.Data != "" {
		history.Data = input.Data
	}

	if err := config.DB.Save(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update history"})
		return
	}

	c.Status(http.StatusOK)
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

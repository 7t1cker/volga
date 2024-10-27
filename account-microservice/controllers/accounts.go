package controllers

import (
	"net/http"
	"strconv"

	"account-microservice/config"
	"account-microservice/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetCurrentAccount(c *gin.Context) {
	accountID := c.GetUint("account_id")

	var account models.Account
	if err := config.DB.Preload("Roles").First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func UpdateCurrentAccount(c *gin.Context) {
	accountID := c.GetUint("account_id")

	var input struct {
		LastName  string `json:"lastName"`
		FirstName string `json:"firstName"`
		Password  string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account models.Account
	if err := config.DB.First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account not found"})
		return
	}

	if input.LastName != "" {
		account.LastName = input.LastName
	}
	if input.FirstName != "" {
		account.FirstName = input.FirstName
	}
	if input.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password encryption failed"})
			return
		}
		account.Password = string(passwordHash)
	}

	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update account"})
		return
	}

	c.Status(http.StatusOK)
}

func GetAllAccounts(c *gin.Context) {
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

	var accounts []models.Account
	query := config.DB.Preload("Roles")

	if fromStr != "" && countStr != "" {
		query = query.Offset(from).Limit(count)
	}

	if err := query.Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

func CreateAccount(c *gin.Context) {
	var input struct {
		LastName  string   `json:"lastName" binding:"required"`
		FirstName string   `json:"firstName" binding:"required"`
		Username  string   `json:"username" binding:"required"`
		Password  string   `json:"password" binding:"required"`
		Roles     []string `json:"roles" binding:"required"`
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

	var roles []*models.Role
	for _, roleName := range input.Roles {
		var role models.Role
		if err := config.DB.FirstOrCreate(&role, models.Role{Name: roleName}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
			return
		}
		roles = append(roles, &role)
	}

	account := models.Account{
		LastName:  input.LastName,
		FirstName: input.FirstName,
		Username:  input.Username,
		Password:  string(passwordHash),
		Roles:     roles,
	}

	if err := config.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.Status(http.StatusCreated)
}

func UpdateAccount(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		LastName  string   `json:"lastName"`
		FirstName string   `json:"firstName"`
		Username  string   `json:"username"`
		Password  string   `json:"password"`
		Roles     []string `json:"roles"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account models.Account
	if err := config.DB.Preload("Roles").First(&account, id).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account not found"})
		return
	}

	if input.LastName != "" {
		account.LastName = input.LastName
	}
	if input.FirstName != "" {
		account.FirstName = input.FirstName
	}
	if input.Username != "" {
		account.Username = input.Username
	}
	if input.Password != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password encryption failed"})
			return
		}
		account.Password = string(passwordHash)
	}
	if len(input.Roles) > 0 {
		var roles []*models.Role
		for _, roleName := range input.Roles {
			var role models.Role
			if err := config.DB.FirstOrCreate(&role, models.Role{Name: roleName}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
				return
			}
			roles = append(roles, &role)
		}
		account.Roles = roles
	}

	if err := config.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&account).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update account"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Where("id = ?", id).Delete(&models.Account{}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete account"})
		return
	}

	c.Status(http.StatusOK)
}

func CheckUserRole(c *gin.Context) {
	accountIDStr := c.Param("id")
	
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil || accountID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var account models.Account

	if err := config.DB.Preload("Roles").First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account not found"})
		return
	}

	if !containsRole2(account.Roles, "user") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func containsRole2(roles []*models.Role, targetRole string) bool {
	for _, role := range roles {
		if role.Name == targetRole {
			return true
		}
	}
	return false
}

func containsRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

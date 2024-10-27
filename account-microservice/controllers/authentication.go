package controllers

import (
	"account-microservice/config"
	"account-microservice/models"
	"account-microservice/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	var input struct {
		LastName  string `json:"lastName" binding:"required"`
		FirstName string `json:"firstName" binding:"required"`
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
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

	var userRole models.Role
	if err := config.DB.FirstOrCreate(&userRole, models.Role{Name: "user"}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or find user role"})
		return
	}

	account := models.Account{
		LastName:  input.LastName,
		FirstName: input.FirstName,
		Username:  input.Username,
		Password:  string(passwordHash),
		Roles:     []*models.Role{&userRole},
	}

	if err := config.DB.Create(&account).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.Status(http.StatusCreated)
}

func SignIn(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account models.Account
	if err := config.DB.Preload("Roles").Where("username = ?", input.Username).First(&account).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	var roles []string
	for _, role := range account.Roles {
		roles = append(roles, role.Name)
	}

	accessToken, err := utils.GenerateAccessToken(account.ID, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(account.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	token := models.Token{
		Token:     refreshToken,
		AccountID: account.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}
	config.DB.Create(&token)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func SignOut(c *gin.Context) {
	accountID := c.GetUint("account_id")
	config.DB.Where("account_id = ?", accountID).Delete(&models.Token{})
	c.Status(http.StatusOK)
}

func ValidateToken(c *gin.Context) {
	accessToken := c.Query("accessToken")
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "accessToken query parameter is required"})
		return
	}

	token, err := utils.ValidateToken(accessToken, os.Getenv("ACCESS_TOKEN_SECRET"))
	isValid := err == nil && token.Valid

	c.JSON(http.StatusOK, gin.H{"isValid": isValid})
}

func RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Refresh token secret not set"})
		return
	}

	token, err := jwt.Parse(input.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_TOKEN_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token or expired",
		})
		return
	}

	accountIDFloat, ok := claims["account_id"].(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims: account_id not found"})
		return
	}
	accountID := uint(accountIDFloat)

	var storedToken models.Token
	if err := config.DB.Where("token = ? AND account_id = ?", input.RefreshToken, accountID).First(&storedToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token in DB", "details": err.Error()})
		return
	}

	var account models.Account
	if err := config.DB.Preload("Roles").First(&account, accountID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Account not found", "details": err.Error()})
		return
	}

	var roles []string
	for _, role := range account.Roles {
		roles = append(roles, role.Name)
	}

	newAccessToken, err := utils.GenerateAccessToken(account.ID, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token", "details": err.Error()})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken(account.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token", "details": err.Error()})
		return
	}

	storedToken.Token = newRefreshToken
	storedToken.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)
	if err := config.DB.Save(&storedToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update refresh token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  newAccessToken,
		"refreshToken": newRefreshToken,
	})
}

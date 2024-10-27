package middlewares

import (
	"net/http"
	"strings"

	"hospital_service/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(accountService *utils.AccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be 'Bearer {token}'"})
			return
		}

		tokenString := parts[1]

		isValid, err := accountService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error validating token", "details": err.Error()})
			return
		}
		if !isValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("accessToken", tokenString)
		c.Next()
	}
}

func AdminMiddleware(accountService *utils.AccountService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, exists := c.Get("accessToken")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access token not found"})
			return
		}

		roles, err := accountService.GetUserRoles(accessToken.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user roles", "details": err.Error()})
			return
		}

		isAdmin := false
		for _, role := range roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			return
		}

		c.Next()
	}
}

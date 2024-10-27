package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("ACCESS_TOKEN_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            return
        }

        c.Set("account_id", uint(claims["account_id"].(float64)))
        c.Set("roles", claims["roles"])
        c.Next()
    }
}

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, exists := c.Get("roles")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Roles not found"})
            return
        }

        rolesSlice := roles.([]interface{})
        isAdmin := false
        for _, role := range rolesSlice {
            if role == "admin" {
                isAdmin = true
                break
            }
        }

        if !isAdmin {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin privileges required"})
            return
        }

        c.Next()
    }
}

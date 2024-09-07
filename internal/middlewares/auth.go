package middlewares

import (
	"avana/internal/config"
	"avana/internal/users"
	"avana/internal/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)


func RequireAuth(c *gin.Context) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        c.JSON(http.StatusUnauthorized, gin.H{"message": utils.ValidateTokenError})
        c.Abort()
        return
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    secret := "chgjfskiuyfgshdigjhv"

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(secret), nil
    })

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"message": utils.ValidateTokenError})
        c.Abort()
        return
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if float64(time.Now().Unix()) > claims["exp"].(float64) {
            c.JSON(http.StatusUnauthorized, gin.H{"message": utils.ValidateTokenError})
            c.Abort()
            return
        }

        var user users.User
        if err := config.DB.First(&user, "email = ?", claims["sub"]).Error;
				err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"message": utils.ValidateTokenError})
					c.Abort()
					return
				}

        if user.ID == 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"message": utils.ValidateTokenError})
            c.Abort()
            return
        }

        c.Set("userID", user.ID)
        c.Next()
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": utils.ValidateTokenError})
        c.Abort()
    }
}
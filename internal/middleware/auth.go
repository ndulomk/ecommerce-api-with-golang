package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header missing",
			})
			return
		}
		
		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use 'Bearer <token>'",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token is empty",
			})
			return
		}

		// Parse e validar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verificar o método de assinatura
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		if !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Token is not valid",
			})
			return
		}

		// Extrair claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Verificar expiração
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Token has expired",
				})
				return
			}
		}

		// Extrair user ID
		userIDStr, err := claims.GetSubject()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token subject",
			})
			return
		}

		if userIDStr == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found in token",
			})
			return
		}

		// Converter para int64
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user ID in token",
			})
			return
		}

		// Extrair role
		role, ok := claims["role"].(string)
		if !ok {
			role = "buyer" 
		}

		// Adicionar informações ao contexto
		ctx.Set("userID", userID)
		ctx.Set("userRole", role)
		
		ctx.Next()
	}
}

// RoleMiddleware verifica se o usuário tem uma das roles permitidas
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole, exists := ctx.Get("userRole")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "User role not found",
			})
			return
		}

		role, ok := userRole.(string)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Invalid user role",
			})
			return
		}

		// Verificar se a role está na lista de permitidas
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
	}
}
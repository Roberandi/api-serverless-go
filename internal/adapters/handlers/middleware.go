package handlers

import (
	"net/http"
	"strings"

	"apigolang/internal/adapters/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware es el guardia de seguridad que intercepta las peticiones
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Buscamos el pase en la cabecera (Header) de la petición
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Acceso denegado: Token requerido"})
			c.Abort() // Abortamos, no pasa al CRUD
			return
		}

		// 2. El token estándar viene en formato "Bearer eyJhbG..."
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Validamos que el token sea real y no haya sido inventado
		err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Si todo está bien, le damos permiso de pasar al CRUD
		c.Next()
	}
}

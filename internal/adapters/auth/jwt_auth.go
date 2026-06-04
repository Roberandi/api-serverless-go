package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("tu_clave_secreta_muy_segura")

func GenerateToken(matricula string) (string, error) {
	// ... tu código actual de GenerateToken se queda igual ...
	claims := jwt.MapClaims{
		"matricula": matricula,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// NUEVO: ValidateToken comprueba que el token sea auténtico y no haya expirado
func ValidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return errors.New("token inválido o expirado")
	}

	return nil
}

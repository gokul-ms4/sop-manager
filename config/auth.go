package config

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwt.MapClaims)
		},
	})
}

func GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

func GetUserID(c echo.Context) int {
	token := c.Get("user").(*jwt.Token)
	claims := token.Claims.(*jwt.MapClaims)
	userID := int((*claims)["user_id"].(float64))
	return userID
}

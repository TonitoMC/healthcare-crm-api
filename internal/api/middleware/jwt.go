package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	authModels "github.com/tonitomc/healthcare-crm-api/internal/domain/auth/models"
)

// JWTMiddleware validates JWT tokens and injects *jwt.Token into context (key "user").
// Claims type is your custom struct via NewClaimsFunc.
func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(secret),
		SigningMethod: echojwt.AlgorithmHS256, // "HS256"
		ContextKey:    "user",                 // c.Get("user") -> *jwt.Token
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(authModels.Claims)
		},
	})
}

package middleware

import (
	"net/http"

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
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/api/auth/login"
		},
	})
}

// RequireAuth ensures a valid JWT exists in context.
// It does NOT check permissions — only authentication.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := c.Get("user").(*jwt.Token)
			if !ok || token == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token no válido o ausente.",
				})
			}

			claims, ok := token.Claims.(*authModels.Claims)
			if !ok || claims == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token inválido.",
				})
			}

			// Optional safety check
			if claims.UserID <= 0 {
				return c.JSON(http.StatusUnauthorized, echo.Map{
					"error": "Token sin ID de usuario válido.",
				})
			}

			return next(c)
		}
	}
}

// GetClaims extracts *authModels.Claims from context.
// Returns nil if unavailable.
func GetClaims(c echo.Context) *authModels.Claims {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return nil
	}

	claims, ok := token.Claims.(*authModels.Claims)
	if !ok {
		return nil
	}

	return claims
}

package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/tonitomc/healthcare-crm-api/internal/domain/user"
	"github.com/tonitomc/healthcare-crm-api/internal/domain/user/models"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// Service handles authentication and authorization logic
type Service struct {
	userService *user.Service
	jwtSecret   []byte
}

// NewService creates a new AuthService
func NewService(userService *user.Service, jwtSecret string) *Service {
	return &Service{
		userService: userService,
		jwtSecret:   []byte(jwtSecret),
	}
}

// Register handles user creation and password hashing
func (s *Service) Register(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return fmt.Errorf("Register: %w", appErr.ErrInvalidInput)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Register (hash): %w", err)
	}

	if err := s.userService.CreateUser(username, email, string(hash)); err != nil {
		return fmt.Errorf("Register: %w", err)
	}

	return nil
}

// Login validates credentials and returns a JWT if successful
func (s *Service) Login(identifier, password string) (string, error) {
	user, err := s.userService.GetByUsernameOrEmail(identifier)
	if err != nil {
		return "", fmt.Errorf("Login: %w", appErr.ErrInvalidCredentials)
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("Login (compare): %w", appErr.ErrInvalidCredentials)
	}

	// Load roles and permissions for JWT payload
	roles, perms, err := s.userService.GetRolesAndPermissions(user.ID)
	if err != nil {
		return "", fmt.Errorf("Login (roles/perms): %w", err)
	}

	token, err := s.generateJWT(user, roles, perms)
	if err != nil {
		return "", fmt.Errorf("Login (token): %w", err)
	}

	return token, nil
}

// generateJWT creates a signed JWT token for the user

func (s *Service) generateJWT(u *models.User, roles []models.Role, perms []models.Permission) (string, error) {
	// Flatten roles and permissions into string slices
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	permNames := make([]string, len(perms))
	for i, p := range perms {
		permNames[i] = p.Name
	}

	// Build claims
	claims := jwt.MapClaims{
		"user_id":     u.ID,
		"username":    u.Username,
		"roles":       roleNames,
		"permissions": permNames,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	// Sign token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("generateJWT: %w", err)
	}

	return signed, nil
}

// ValidateToken verifies and parses a JWT
func (s *Service) ValidateToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateToken: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("ValidateToken: %w", appErr.ErrInvalidToken)
	}

	return token, nil
}

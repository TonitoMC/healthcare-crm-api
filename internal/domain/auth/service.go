// internal/domain/auth/service.go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	authModels "github.com/tonitomc/healthcare-crm-api/internal/domain/auth/models"
	rbacDomain "github.com/tonitomc/healthcare-crm-api/internal/domain/rbac"
	rbacModels "github.com/tonitomc/healthcare-crm-api/internal/domain/rbac/models"
	userDomain "github.com/tonitomc/healthcare-crm-api/internal/domain/user"
	appErr "github.com/tonitomc/healthcare-crm-api/pkg/errors"
)

// -----------------------------------------------------------------------------
// Service Interface
// -----------------------------------------------------------------------------

type Service interface {
	Register(username, email, password string) error
	Login(identifier, password string) (string, error)
	ValidateToken(tokenStr string) (*jwt.Token, *authModels.Claims, error)
	ChangePassword(userID int, oldPassword, newPassword string) error
}

// -----------------------------------------------------------------------------
// Implementation
// -----------------------------------------------------------------------------

type service struct {
	userService userDomain.Service
	rbacService rbacDomain.Service
	jwtSecret   []byte
	accessTTL   time.Duration
	issuer      string
}

// Config allows customizing the Auth service behavior.
type Config struct {
	JWTSecret string
	AccessTTL time.Duration
	Issuer    string
}

// NewService constructs a new Auth service.
func NewService(userSvc userDomain.Service, rbacSvc rbacDomain.Service, cfg Config) Service {
	if cfg.AccessTTL == 0 {
		cfg.AccessTTL = 24 * time.Hour
	}
	return &service{
		userService: userSvc,
		rbacService: rbacSvc,
		jwtSecret:   []byte(cfg.JWTSecret),
		accessTTL:   cfg.AccessTTL,
		issuer:      cfg.Issuer,
	}
}

// -----------------------------------------------------------------------------
// Register / Login / Validate
// -----------------------------------------------------------------------------

func (s *service) Register(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return appErr.Wrap("AuthService.Register", appErr.ErrInvalidInput, nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return appErr.Wrap("AuthService.Register(hash)", appErr.ErrInternal, err)
	}

	if err := s.userService.CreateUser(username, email, string(hash)); err != nil {
		return err // already wrapped
	}
	return nil
}

func (s *service) Login(identifier, password string) (string, error) {
	if identifier == "" || password == "" {
		return "", appErr.Wrap("AuthService.Login", appErr.ErrInvalidInput, nil)
	}

	u, err := s.userService.GetByUsernameOrEmail(identifier)
	if err != nil {
		return "", appErr.Wrap("AuthService.Login(user lookup)", appErr.ErrInvalidCredentials, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", appErr.Wrap("AuthService.Login(compare)", appErr.ErrInvalidCredentials, err)
	}

	rbacCtx, err := s.rbacService.GetUserAccess(u.ID)
	if err != nil {
		return "", appErr.Wrap("AuthService.Login(rbac)", appErr.ErrInternal, err)
	}

	token, err := s.generateJWT(rbacCtx)
	if err != nil {
		return "", appErr.Wrap("AuthService.Login(token)", appErr.ErrInternal, err)
	}

	return token, nil
}

func (s *service) ValidateToken(tokenStr string) (*jwt.Token, *authModels.Claims, error) {
	if tokenStr == "" {
		return nil, nil, appErr.Wrap("AuthService.ValidateToken", appErr.ErrInvalidToken, nil)
	}

	claims := &authModels.Claims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	token, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, nil, appErr.Wrap("AuthService.ValidateToken(parse)", appErr.ErrInvalidToken, err)
	}
	if !token.Valid {
		return nil, nil, appErr.Wrap("AuthService.ValidateToken", appErr.ErrInvalidToken, nil)
	}
	if s.issuer != "" && claims.Issuer != s.issuer {
		return nil, nil, appErr.Wrap("AuthService.ValidateToken(issuer)", appErr.ErrInvalidToken, nil)
	}

	return token, claims, nil
}

// -----------------------------------------------------------------------------
// JWT generator
// -----------------------------------------------------------------------------

func (s *service) generateJWT(rbacCtx *rbacModels.RBAC) (string, error) {
	roleNames := make([]string, 0, len(rbacCtx.Roles))
	for _, r := range rbacCtx.Roles {
		roleNames = append(roleNames, r.Name)
	}

	permNames := make([]string, 0, len(rbacCtx.Permissions))
	for _, p := range rbacCtx.Permissions {
		permNames = append(permNames, p.Name)
	}

	now := time.Now()
	claims := authModels.Claims{
		UserID:      rbacCtx.User.ID,
		Username:    rbacCtx.User.Username,
		Roles:       roleNames,
		Permissions: permNames,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", appErr.Wrap("AuthService.generateJWT", appErr.ErrInternal, err)
	}

	return signed, nil
}

func (s *service) ChangePassword(userID int, oldPassword, newPassword string) error {
	if userID <= 0 || oldPassword == "" || newPassword == "" {
		return appErr.Wrap("AuthService.ChangePassword", appErr.ErrInvalidInput, nil)
	}

	u, err := s.userService.GetByID(userID)
	if err != nil {
		return appErr.Wrap("AuthService.ChangePassword(user)", appErr.ErrInvalidCredentials, err)
	}

	// verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return appErr.Wrap("AuthService.ChangePassword(compare)", appErr.ErrInvalidCredentials, err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return appErr.Wrap("AuthService.ChangePassword(hash)", appErr.ErrInternal, err)
	}

	u.PasswordHash = string(hashed)

	if err := s.userService.UpdateUser(u); err != nil {
		return err
	}

	return nil
}

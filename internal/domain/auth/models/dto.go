package models

// RegisterRequest represents the expected body for /auth/register
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the expected body for /auth/login
type LoginRequest struct {
	Identifier string `json:"identifier"` // username or email
	Password   string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

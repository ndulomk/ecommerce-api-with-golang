package models

import (
	"time"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username" validate:"required,alphanum,min=3,max=100"`
	Email        string    `db:"email" json:"email" validate:"required,email,max=255"`
	Phone        *string   `db:"phone" json:"phone,omitempty" validate:"omitempty,e164"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role" validate:"omitempty,oneof=buyer seller admin supplier"`
	Status       string    `db:"status" json:"status" validate:"omitempty,oneof=active suspended banned"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Validate user struct
func (u *User) Validate() error {
	return validate.Struct(u)
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,alphanum,min=3,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Phone    string `json:"phone" validate:"omitempty,e164"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

// Validate register request
func (r *RegisterRequest) Validate() error {
	return validate.Struct(r)
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Validate login request
func (l *LoginRequest) Validate() error {
	return validate.Struct(l)
}

type UpdateUserRequest struct {
	Username *string `json:"username" validate:"omitempty,alphanum,min=3,max=100"`
	Email    *string `json:"email" validate:"omitempty,email,max=255"`
	Phone    *string `json:"phone" validate:"omitempty,e164"`
}

// Validate update request
func (u *UpdateUserRequest) Validate() error {
	return validate.Struct(u)
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

// UserProfile representa o perfil público do usuário 
type UserProfile struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToProfile converte User para UserProfile
func (u *User) ToProfile() UserProfile {
	return UserProfile{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"modress/internal/models"
	"modress/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserData    = errors.New("invalid user data")
)

type AuthService interface {
	Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, *models.User, error)
	GetUser(ctx context.Context, id int64) (*models.User, error)
	UpdateUser(ctx context.Context, id int64, req models.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type authService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUserData, err)
	}

	// Verificar se o email já existe
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Criar novo usuário
	newUser := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		Phone:        s.stringToPointer(req.Phone),
		PasswordHash: string(hashedPassword),
		Role:         "buyer",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validar struct
	if err := newUser.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUserData, err)
	}

	// Criar no banco
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return newUser, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	// Validar entrada
	if email == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	// Buscar usuário por email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return "", nil, ErrInvalidCredentials
	}

	// Verificar se o usuário está ativo
	if user.Status != "active" {
		return "", nil, ErrInvalidCredentials
	}

	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Gerar token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  fmt.Sprintf("%d", user.ID),
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("error generating token: %w", err)
	}

	return tokenString, user, nil
}

func (s *authService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	if id <= 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *authService) UpdateUser(ctx context.Context, id int64, req models.UpdateUserRequest) (*models.User, error) {
	if id <= 0 {
		return nil, ErrUserNotFound
	}

	// Validar request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUserData, err)
	}

	// Buscar usuário existente
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Verificar se email já existe 
	if req.Email != nil && *req.Email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil {
			return nil, fmt.Errorf("error checking email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}
	}

	// Atualizar campos permitidos
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}

	user.UpdatedAt = time.Now()

	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUserData, err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return user, nil
}

func (s *authService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrUserNotFound
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("error deleting user: %w", err)
	}
	return nil
}

// Helper function para converter string em pointer
func (s *authService) stringToPointer(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}
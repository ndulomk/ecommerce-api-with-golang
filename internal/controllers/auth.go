package controllers

import (
	"errors"
	"log"
	"modress/internal/models"
	"modress/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register cria um novo usuário
func (c *AuthController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.handleValidationError(ctx, err)
		return
	}

	user, err := c.authService.Register(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		if errors.Is(err, services.ErrInvalidUserData) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		log.Printf("Error registering user: %v", err)
		return
	}

	// Retornar apenas o perfil público
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user.ToProfile(),
	})
}

// Login autentica um usuário
func (c *AuthController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.handleValidationError(ctx, err)
		return
	}

	token, user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, models.LoginResponse{
		AccessToken: token,
		User:        *user,
	})
}

// GetProfile retorna o perfil do usuário logado
func (c *AuthController) GetProfile(ctx *gin.Context) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := c.authService.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, user.ToProfile())
}

// UpdateProfile atualiza o perfil do usuário logado
func (c *AuthController) UpdateProfile(ctx *gin.Context) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.handleValidationError(ctx, err)
		return
	}

	user, err := c.authService.UpdateUser(ctx.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if errors.Is(err, services.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		if errors.Is(err, services.ErrInvalidUserData) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user.ToProfile(),
	})
}

// DeleteProfile deleta o perfil do usuário logado
func (c *AuthController) DeleteProfile(ctx *gin.Context) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.authService.DeleteUser(ctx.Request.Context(), userID); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

// getUserIDFromContext extrai o ID do usuário do contexto
func (c *AuthController) getUserIDFromContext(ctx *gin.Context) (int64, error) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return 0, errors.New("user ID not found in context")
	}

	id, ok := userID.(int64)
	if !ok {
		return 0, errors.New("invalid user ID type")
	}

	return id, nil
}

// handleValidationError trata erros de validação
func (c *AuthController) handleValidationError(ctx *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		var errorMessages []string
		for _, fieldError := range ve {
			errorMessages = append(errorMessages, c.getValidationErrorMessage(fieldError))
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": errorMessages,
		})
		return
	}

	ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
}

// getValidationErrorMessage retorna mensagem de erro personalizada
func (c *AuthController) getValidationErrorMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + fe.Param() + " characters long"
	case "max":
		return field + " must be at most " + fe.Param() + " characters long"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "e164":
		return field + " must be a valid phone number in E164 format"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	default:
		return field + " is invalid"
	}
}
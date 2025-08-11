package models

import (
	"strings"
	"time"
)

type Store struct {
	ID          int64     `db:"id" json:"id"`
	OwnerID     int64     `db:"owner_id" json:"owner_id"`
	Name        string    `db:"name" json:"name" validate:"required,min=3,max=150"`
	Slug        string    `db:"slug" json:"slug" validate:"required,min=3,max=150,slug"`
	Description *string   `db:"description" json:"description,omitempty"`
	LogoURL     *string   `db:"logo_url" json:"logo_url,omitempty" validate:"omitempty,url"`
	BannerURL   *string   `db:"banner_url" json:"banner_url,omitempty" validate:"omitempty,url"`
	IsApproved  bool      `db:"is_approved" json:"is_approved"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Validate store struct
func (s *Store) Validate() error {
	return validate.Struct(s)
}

// GenerateSlug gera um slug a partir do nome
func (s *Store) GenerateSlug() {
	slug := strings.ToLower(s.Name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s.Slug = result.String()
}

type CreateStoreRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=150"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty" validate:"omitempty,url"`
	BannerURL   *string `json:"banner_url,omitempty" validate:"omitempty,url"`
}

// Validate create store request
func (r *CreateStoreRequest) Validate() error {
	return validate.Struct(r)
}

type UpdateStoreRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=150"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty" validate:"omitempty,url"`
	BannerURL   *string `json:"banner_url,omitempty" validate:"omitempty,url"`
}

// Validate update store request
func (r *UpdateStoreRequest) Validate() error {
	return validate.Struct(r)
}

type StoreResponse struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	LogoURL     *string   `json:"logo_url,omitempty"`
	BannerURL   *string   `json:"banner_url,omitempty"`
	IsApproved  bool      `json:"is_approved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converte Store para StoreResponse
func (s *Store) ToResponse() StoreResponse {
	return StoreResponse{
		ID:          s.ID,
		OwnerID:     s.OwnerID,
		Name:        s.Name,
		Slug:        s.Slug,
		Description: s.Description,
		LogoURL:     s.LogoURL,
		BannerURL:   s.BannerURL,
		IsApproved:  s.IsApproved,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

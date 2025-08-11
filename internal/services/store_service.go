package services

import (
	"context"
	"database/sql"
	"fmt"
	"modress/internal/models"
	"modress/internal/repositories"
	"time"
)

// StoreService interface
type StoreService interface {
	CreateStore(ctx context.Context, ownerID int64, req *models.CreateStoreRequest) (*models.StoreResponse, error)
	GetStoreByID(ctx context.Context, id int64) (*models.StoreResponse, error)
	GetStoreBySlug(ctx context.Context, slug string) (*models.StoreResponse, error)
	GetStoreByOwnerID(ctx context.Context, ownerID int64) (*models.StoreResponse, error)
	UpdateStore(ctx context.Context, id int64, ownerID int64, req *models.UpdateStoreRequest) (*models.StoreResponse, error)
	DeleteStore(ctx context.Context, id int64, ownerID int64) error
	ListStores(ctx context.Context, page, limit int) ([]models.StoreResponse, error)
	ListApprovedStores(ctx context.Context, page, limit int) ([]models.StoreResponse, error)
	ApproveStore(ctx context.Context, id int64) error
}

type storeService struct {
	storeRepo repositories.StoreRepository
}

func NewStoreService(storeRepo repositories.StoreRepository) StoreService {
	return &storeService{
		storeRepo: storeRepo,
	}
}

func (s *storeService) CreateStore(ctx context.Context, ownerID int64, req *models.CreateStoreRequest) (*models.StoreResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Verificar se o usuário já possui uma loja
	existingStore, err := s.storeRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error checking existing store: %w", err)
	}
	if existingStore != nil {
		return nil, fmt.Errorf("user already has a store")
	}

	now := time.Now()
	store := &models.Store{
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		BannerURL:   req.BannerURL,
		IsApproved:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Gerar slug automaticamente
	store.GenerateSlug()

	// Verificar se o slug já existe
	existingSlug, err := s.storeRepo.FindBySlug(ctx, store.Slug)
	if err != nil {
		return nil, fmt.Errorf("error checking slug: %w", err)
	}
	if existingSlug != nil {
		// Adicionar timestamp ao slug para torná-lo único
		store.Slug = fmt.Sprintf("%s-%d", store.Slug, time.Now().Unix())
	}

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, fmt.Errorf("error creating store: %w", err)
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *storeService) GetStoreByID(ctx context.Context, id int64) (*models.StoreResponse, error) {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *storeService) GetStoreBySlug(ctx context.Context, slug string) (*models.StoreResponse, error) {
	store, err := s.storeRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *storeService) GetStoreByOwnerID(ctx context.Context, ownerID int64) (*models.StoreResponse, error) {
	store, err := s.storeRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *storeService) UpdateStore(ctx context.Context, id int64, ownerID int64, req *models.UpdateStoreRequest) (*models.StoreResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	// Verificar se o usuário é o dono da loja
	if store.OwnerID != ownerID {
		return nil, fmt.Errorf("unauthorized: not store owner")
	}

	// Atualizar apenas os campos fornecidos
	if req.Name != nil {
		store.Name = *req.Name
		store.GenerateSlug() 
		
		// Verificar se o novo slug já existe
		existingSlug, err := s.storeRepo.FindBySlug(ctx, store.Slug)
		if err != nil {
			return nil, fmt.Errorf("error checking slug: %w", err)
		}
		if existingSlug != nil && existingSlug.ID != store.ID {
			store.Slug = fmt.Sprintf("%s-%d", store.Slug, time.Now().Unix())
		}
	}
	if req.Description != nil {
		store.Description = req.Description
	}
	if req.LogoURL != nil {
		store.LogoURL = req.LogoURL
	}
	if req.BannerURL != nil {
		store.BannerURL = req.BannerURL
	}

	store.UpdatedAt = time.Now()

	if err := s.storeRepo.Update(ctx, store); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("store not found")
		}
		return nil, fmt.Errorf("error updating store: %w", err)
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *storeService) DeleteStore(ctx context.Context, id int64, ownerID int64) error {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return fmt.Errorf("store not found")
	}

	// Verificar se o usuário é o dono da loja
	if store.OwnerID != ownerID {
		return fmt.Errorf("unauthorized: not store owner")
	}

	if err := s.storeRepo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("store not found")
		}
		return fmt.Errorf("error deleting store: %w", err)
	}

	return nil
}

func (s *storeService) ListStores(ctx context.Context, page, limit int) ([]models.StoreResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stores, err := s.storeRepo.List(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing stores: %w", err)
	}

	responses := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		responses[i] = store.ToResponse()
	}

	return responses, nil
}

func (s *storeService) ListApprovedStores(ctx context.Context, page, limit int) ([]models.StoreResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stores, err := s.storeRepo.ListApproved(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing approved stores: %w", err)
	}

	responses := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		responses[i] = store.ToResponse()
	}

	return responses, nil
}

func (s *storeService) ApproveStore(ctx context.Context, id int64) error {
	if err := s.storeRepo.ApproveStore(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("store not found")
		}
		return fmt.Errorf("error approving store: %w", err)
	}

	return nil
}
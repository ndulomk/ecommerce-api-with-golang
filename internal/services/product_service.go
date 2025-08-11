package services

import (
	"context"
	"database/sql"
	"fmt"
	"modress/internal/models"
	"modress/internal/repositories"
	"time"
)

// ProductService interface
type ProductService interface {
	CreateProduct(ctx context.Context, storeID int64, req *models.CreateProductRequest) (*models.ProductResponse, error)
	GetProductByID(ctx context.Context, id int64) (*models.ProductResponse, error)
	GetProductsByStoreID(ctx context.Context, storeID int64, page, limit int) ([]models.ProductResponse, error)
	GetProductsByCategory(ctx context.Context, category string, page, limit int) ([]models.ProductResponse, error)
	UpdateProduct(ctx context.Context, id int64, storeID int64, req *models.UpdateProductRequest) (*models.ProductResponse, error)
	DeleteProduct(ctx context.Context, id int64, storeID int64) error
	ListProducts(ctx context.Context, page, limit int) ([]models.ProductResponse, error)
	SearchProducts(ctx context.Context, query string, page, limit int) ([]models.ProductResponse, error)
	UpdateProductQuantity(ctx context.Context, id int64, storeID int64, quantity int) error
	GetStoreByOwnerID(ctx context.Context, ownerID int64) (*models.StoreResponse, error) 
	CreateProductImage(ctx context.Context, productID, storeID int64, image *models.ProductImage) error
	GetProductImages(ctx context.Context, productID int64) ([]models.ProductImage, error) 
}

type productService struct {
	productRepo repositories.ProductRepository
	storeRepo   repositories.StoreRepository
}

func NewProductService(productRepo repositories.ProductRepository, storeRepo repositories.StoreRepository) ProductService {
	return &productService{
		productRepo: productRepo,
		storeRepo:   storeRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, storeID int64, req *models.CreateProductRequest) (*models.ProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Verificar se a loja existe
	store, err := s.storeRepo.FindByID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("error finding store: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	now := time.Now()
	product := &models.Product{
		StoreID:     storeID,
		Title:       req.Title,
		Description: req.Description,
		PriceCents:  int(req.Price * 100), 
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		Quantity:    req.Quantity,
		IsActive:    true,
		Category:    req.Category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.Cost != nil {
		costCents := int(*req.Cost * 100)
		product.CostCents = &costCents
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *productService) GetProductByID(ctx context.Context, id int64) (*models.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *productService) GetProductsByStoreID(ctx context.Context, storeID int64, page, limit int) ([]models.ProductResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := s.productRepo.FindByStoreID(ctx, storeID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error finding products by store: %w", err)
	}

	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

func (s *productService) GetProductsByCategory(ctx context.Context, category string, page, limit int) ([]models.ProductResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := s.productRepo.FindByCategory(ctx, category, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error finding products by category: %w", err)
	}

	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id int64, storeID int64, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding product: %w", err)
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}

	// Verificar se o produto pertence à loja
	if product.StoreID != storeID {
		return nil, fmt.Errorf("unauthorized: product does not belong to store")
	}

	// Atualizar apenas os campos fornecidos
	if req.Title != nil {
		product.Title = *req.Title
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.Price != nil {
		product.PriceCents = int(*req.Price * 100)
	}
	if req.Cost != nil {
		costCents := int(*req.Cost * 100)
		product.CostCents = &costCents
	}
	if req.SKU != nil {
		product.SKU = req.SKU
	}
	if req.Barcode != nil {
		product.Barcode = req.Barcode
	}
	if req.Quantity != nil {
		product.Quantity = *req.Quantity
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}
	if req.Category != nil {
		product.Category = req.Category
	}

	product.UpdatedAt = time.Now()

	if err := s.productRepo.Update(ctx, product); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *productService) GetStoreByOwnerID(ctx context.Context, ownerID int64) (*models.StoreResponse, error) {
	store, err := s.storeRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error finding store by owner ID: %w", err)
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}

	response := store.ToResponse()
	return &response, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id int64, storeID int64) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error finding product: %w", err)
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// Verificar se o produto pertence à loja
	if product.StoreID != storeID {
		return fmt.Errorf("unauthorized: product does not belong to store")
	}

	if err := s.productRepo.Delete(ctx, id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

func (s *productService) ListProducts(ctx context.Context, page, limit int) ([]models.ProductResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := s.productRepo.List(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}

	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

func (s *productService) SearchProducts(ctx context.Context, query string, page, limit int) ([]models.ProductResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	products, err := s.productRepo.Search(ctx, query, page, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching products: %w", err)
	}

	responses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return responses, nil
}

func (s *productService) UpdateProductQuantity(ctx context.Context, id int64, storeID int64, quantity int) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error finding product: %w", err)
	}
	if product == nil {
		return fmt.Errorf("product not found")
	}

	// Verificar se o produto pertence à loja
	if product.StoreID != storeID {
		return fmt.Errorf("unauthorized: product does not belong to store")
	}

	if quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}

	if err := s.productRepo.UpdateQuantity(ctx, id, quantity); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("error updating product quantity: %w", err)
	}

	return nil
}


func (s *productService) CreateProductImage(ctx context.Context, productID, storeID int64, image *models.ProductImage) error {
    // Validate image
    if err := image.Validate(); err != nil {
        return fmt.Errorf("validation error: %w", err)
    }

    // Verify product exists and belongs to the store
    product, err := s.productRepo.FindByID(ctx, productID)
    if err != nil {
        return fmt.Errorf("error finding product: %w", err)
    }
    if product == nil {
        return fmt.Errorf("product not found")
    }
    if product.StoreID != storeID {
        return fmt.Errorf("unauthorized: product does not belong to store")
    }

    // Set product ID and creation time
    image.ProductID = productID
    image.CreatedAt = time.Now()

    // Save image metadata to database
    if err := s.productRepo.CreateImage(ctx, image); err != nil {
        return fmt.Errorf("error creating product image: %w", err)
    }

    return nil
}

func (s *productService) GetProductImages(ctx context.Context, productID int64) ([]models.ProductImage, error) {
    images, err := s.productRepo.FindImagesByProductID(ctx, productID)
    if err != nil {
        return nil, fmt.Errorf("error retrieving product images: %w", err)
    }
    return images, nil
}
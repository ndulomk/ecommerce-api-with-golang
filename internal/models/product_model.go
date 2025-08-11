	package models

	import (
		"time"
	)
	// Product models
	type Product struct {
		ID          int64     `db:"id" json:"id"`
		StoreID     int64     `db:"store_id" json:"store_id"`
		Title       string    `db:"title" json:"title" validate:"required,min=3,max=255"`
		Description *string   `db:"description" json:"description,omitempty"`
		PriceCents  int       `db:"price_cents" json:"price_cents" validate:"min=0"`
		CostCents   *int      `db:"cost_cents" json:"cost_cents,omitempty" validate:"omitempty,min=0"`
		SKU         *string   `db:"sku" json:"sku,omitempty" validate:"omitempty,max=100"`
		Barcode     *string   `db:"barcode" json:"barcode,omitempty" validate:"omitempty,max=100"`
		Quantity    int       `db:"quantity" json:"quantity" validate:"min=0"`
		IsActive    bool      `db:"is_active" json:"is_active"`
		Category    *string   `db:"category" json:"category,omitempty" validate:"omitempty,max=100"`
		CreatedAt   time.Time `db:"created_at" json:"created_at"`
		UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	}

	// Validate product struct
	func (p *Product) Validate() error {
		return validate.Struct(p)
	}

	// GetPrice retorna o preço em formato decimal
	func (p *Product) GetPrice() float64 {
		return float64(p.PriceCents) / 100
	}

	// SetPrice define o preço a partir de um valor decimal
	func (p *Product) SetPrice(price float64) {
		p.PriceCents = int(price * 100)
	}

	type CreateProductRequest struct {
		Title       string  `json:"title" validate:"required,min=3,max=255"`
		Description *string `json:"description,omitempty"`
		Price       float64 `json:"price" validate:"min=0"`
		Cost        *float64 `json:"cost,omitempty" validate:"omitempty,min=0"`
		SKU         *string `json:"sku,omitempty" validate:"omitempty,max=100"`
		Barcode     *string `json:"barcode,omitempty" validate:"omitempty,max=100"`
		Quantity    int     `json:"quantity" validate:"min=0"`
		Category    *string `json:"category,omitempty" validate:"omitempty,max=100"`
	}

	// Validate create product request
	func (r *CreateProductRequest) Validate() error {
		return validate.Struct(r)
	}

	type UpdateProductRequest struct {
		Title       *string  `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
		Description *string  `json:"description,omitempty"`
		Price       *float64 `json:"price,omitempty" validate:"omitempty,min=0"`
		Cost        *float64 `json:"cost,omitempty" validate:"omitempty,min=0"`
		SKU         *string  `json:"sku,omitempty" validate:"omitempty,max=100"`
		Barcode     *string  `json:"barcode,omitempty" validate:"omitempty,max=100"`
		Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,min=0"`
		IsActive    *bool    `json:"is_active,omitempty"`
		Category    *string  `json:"category,omitempty" validate:"omitempty,max=100"`
	}

	// Validate update product request
	func (r *UpdateProductRequest) Validate() error {
		return validate.Struct(r)
	}

	type ProductResponse struct {
		ID          int64     `json:"id"`
		StoreID     int64     `json:"store_id"`
		Title       string    `json:"title"`
		Description *string   `json:"description,omitempty"`
		Price       float64   `json:"price"`
		Cost        *float64  `json:"cost,omitempty"`
		SKU         *string   `json:"sku,omitempty"`
		Barcode     *string   `json:"barcode,omitempty"`
		Quantity    int       `json:"quantity"`
		IsActive    bool      `json:"is_active"`
		Category    *string   `json:"category,omitempty"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	// ToResponse converte Product para ProductResponse
	func (p *Product) ToResponse() ProductResponse {
		var cost *float64
		if p.CostCents != nil {
			costValue := float64(*p.CostCents) / 100
			cost = &costValue
		}

		return ProductResponse{
			ID:          p.ID,
			StoreID:     p.StoreID,
			Title:       p.Title,
			Description: p.Description,
			Price:       p.GetPrice(),
			Cost:        cost,
			SKU:         p.SKU,
			Barcode:     p.Barcode,
			Quantity:    p.Quantity,
			IsActive:    p.IsActive,
			Category:    p.Category,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	// ProductImage model
	type ProductImage struct {
		ID        int64     `db:"id" json:"id"`
		ProductID int64     `db:"product_id" json:"product_id"`
		URL       string    `db:"url" json:"url" validate:"required,url"`
		AltText   *string   `db:"alt_text" json:"alt_text,omitempty" validate:"omitempty,max=255"`
		Position  int       `db:"position" json:"position" validate:"min=1"`
		IsPrimary bool      `db:"is_primary" json:"is_primary"`
		CreatedAt time.Time `db:"created_at" json:"created_at"`
	}

	// Validate product image struct
	func (pi *ProductImage) Validate() error {
		return validate.Struct(pi)
	}
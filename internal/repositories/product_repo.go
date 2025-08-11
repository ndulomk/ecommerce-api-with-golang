package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"modress/internal/models"

	"github.com/jmoiron/sqlx"
)

// ProductRepository interface
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	FindByID(ctx context.Context, id int64) (*models.Product, error)
	FindByStoreID(ctx context.Context, storeID int64, page, limit int) ([]models.Product, error)
	FindByCategory(ctx context.Context, category string, page, limit int) ([]models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, limit int) ([]models.Product, error)
	Search(ctx context.Context, query string, page, limit int) ([]models.Product, error)
	UpdateQuantity(ctx context.Context, id int64, quantity int) error
	CreateImage(ctx context.Context, image *models.ProductImage) error 
	FindImagesByProductID(ctx context.Context, productID int64) ([]models.ProductImage, error) 

}

type productRepo struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepo{db: db}
}



func (r *productRepo) Create(ctx context.Context, product *models.Product) error {
	query := `
	INSERT INTO products (
		store_id, title, description, price_cents, cost_cents, sku, barcode, 
		quantity, is_active, category, created_at, updated_at
	) VALUES (
		:store_id, :title, :description, :price_cents, :cost_cents, :sku, :barcode,
		:quantity, :is_active, :category, :created_at, :updated_at
	)
	RETURNING id`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	return stmt.GetContext(ctx, &product.ID, product)
}

func (r *productRepo) FindByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `SELECT * FROM products WHERE id = $1`
	var product models.Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &product, err
}

func (r *productRepo) FindByStoreID(ctx context.Context, storeID int64, page, limit int) ([]models.Product, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM products WHERE store_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var products []models.Product
	err := r.db.SelectContext(ctx, &products, query, storeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding products by store: %w", err)
	}

	return products, nil
}

func (r *productRepo) FindByCategory(ctx context.Context, category string, page, limit int) ([]models.Product, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM products WHERE category = $1 AND is_active = true ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	var products []models.Product
	err := r.db.SelectContext(ctx, &products, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding products by category: %w", err)
	}

	return products, nil
}

func (r *productRepo) Update(ctx context.Context, product *models.Product) error {
	query := `
	UPDATE products SET
		title = :title,
		description = :description,
		price_cents = :price_cents,
		cost_cents = :cost_cents,
		sku = :sku,
		barcode = :barcode,
		quantity = :quantity,
		is_active = :is_active,
		category = :category,
		updated_at = :updated_at
	WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepo) List(ctx context.Context, page, limit int) ([]models.Product, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM products WHERE is_active = true ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var products []models.Product
	err := r.db.SelectContext(ctx, &products, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}

	return products, nil
}

func (r *productRepo) Search(ctx context.Context, searchQuery string, page, limit int) ([]models.Product, error) {
	offset := (page - 1) * limit
	query := `
	SELECT * FROM products 
	WHERE is_active = true 
	AND (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%' OR category ILIKE '%' || $1 || '%')
	ORDER BY created_at DESC 
	LIMIT $2 OFFSET $3`

	var products []models.Product
	err := r.db.SelectContext(ctx, &products, query, searchQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error searching products: %w", err)
	}

	return products, nil
}

func (r *productRepo) UpdateQuantity(ctx context.Context, id int64, quantity int) error {
	query := `UPDATE products SET quantity = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, quantity, id)
	if err != nil {
		return fmt.Errorf("error updating product quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *productRepo) CreateImage(ctx context.Context, image *models.ProductImage) error {
    query := `
    INSERT INTO product_images (
        product_id, url, alt_text, position, is_primary, created_at
    ) VALUES (
        :product_id, :url, :alt_text, :position, :is_primary, :created_at
    )
    RETURNING id`

    stmt, err := r.db.PrepareNamedContext(ctx, query)
    if err != nil {
        return fmt.Errorf("error preparing query: %w", err)
    }
    defer stmt.Close()

    return stmt.GetContext(ctx, &image.ID, image)
}

func (r *productRepo) FindImagesByProductID(ctx context.Context, productID int64) ([]models.ProductImage, error) {
    query := `SELECT * FROM product_images WHERE product_id = $1 ORDER BY position ASC`
    var images []models.ProductImage
    err := r.db.SelectContext(ctx, &images, query, productID)
    if err != nil {
        return nil, fmt.Errorf("error finding product images: %w", err)
    }
    return images, nil
}
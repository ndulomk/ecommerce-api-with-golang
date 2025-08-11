package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"modress/internal/models"

	"github.com/jmoiron/sqlx"
)

// StoreRepository interface
type StoreRepository interface {
	Create(ctx context.Context, store *models.Store) error
	FindByID(ctx context.Context, id int64) (*models.Store, error)
	FindByOwnerID(ctx context.Context, ownerID int64) (*models.Store, error)
	FindBySlug(ctx context.Context, slug string) (*models.Store, error)
	Update(ctx context.Context, store *models.Store) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, limit int) ([]models.Store, error)
	ListApproved(ctx context.Context, page, limit int) ([]models.Store, error)
	ApproveStore(ctx context.Context, id int64) error
}

type storeRepo struct {
	db *sqlx.DB
}

func NewStoreRepository(db *sqlx.DB) StoreRepository {
	return &storeRepo{db: db}
}

func (r *storeRepo) Create(ctx context.Context, store *models.Store) error {
	query := `
	INSERT INTO stores (
		owner_id, name, slug, description, logo_url, banner_url, is_approved, created_at, updated_at
	) VALUES (
		:owner_id, :name, :slug, :description, :logo_url, :banner_url, :is_approved, :created_at, :updated_at
	)
	RETURNING id`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	return stmt.GetContext(ctx, &store.ID, store)
}

func (r *storeRepo) FindByID(ctx context.Context, id int64) (*models.Store, error) {
	query := `SELECT * FROM stores WHERE id = $1`
	var store models.Store
	err := r.db.GetContext(ctx, &store, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &store, err
}

func (r *storeRepo) FindByOwnerID(ctx context.Context, ownerID int64) (*models.Store, error) {
	query := `SELECT * FROM stores WHERE owner_id = $1`
	var store models.Store
	err := r.db.GetContext(ctx, &store, query, ownerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &store, err
}

func (r *storeRepo) FindBySlug(ctx context.Context, slug string) (*models.Store, error) {
	query := `SELECT * FROM stores WHERE slug = $1`
	var store models.Store
	err := r.db.GetContext(ctx, &store, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &store, err
}

func (r *storeRepo) Update(ctx context.Context, store *models.Store) error {
	query := `
	UPDATE stores SET
		name = :name,
		slug = :slug,
		description = :description,
		logo_url = :logo_url,
		banner_url = :banner_url,
		is_approved = :is_approved,
		updated_at = :updated_at
	WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, store)
	if err != nil {
		return fmt.Errorf("error updating store: %w", err)
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

func (r *storeRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM stores WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting store: %w", err)
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

func (r *storeRepo) List(ctx context.Context, page, limit int) ([]models.Store, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM stores ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var stores []models.Store
	err := r.db.SelectContext(ctx, &stores, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing stores: %w", err)
	}

	return stores, nil
}

func (r *storeRepo) ListApproved(ctx context.Context, page, limit int) ([]models.Store, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM stores WHERE is_approved = true ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	var stores []models.Store
	err := r.db.SelectContext(ctx, &stores, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing approved stores: %w", err)
	}

	return stores, nil
}

func (r *storeRepo) ApproveStore(ctx context.Context, id int64) error {
	query := `UPDATE stores SET is_approved = true, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error approving store: %w", err)
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

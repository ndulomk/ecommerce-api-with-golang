package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"modress/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, limit int) ([]models.User, error)
	ListAll(ctx context.Context) ([]models.User, error)
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (
		username, email, phone, password_hash, role, status, created_at, updated_at
	) VALUES (
		:username, :email, :phone, :password_hash, :role, :status, :created_at, :updated_at
	)
	RETURNING id`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	return stmt.GetContext(ctx, &user.ID, user)
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT * FROM users WHERE email = $1`
	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users SET
		username = :username,
		email = :email,
		phone = :phone,
		role = :role,
		status = :status,
		updated_at = :updated_at
	WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
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

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
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
func getString(s *string) string {
	if s == nil {
		return "(null)"
	}
	return *s
}


func (r *userRepo) ListAll(ctx context.Context) ([]models.User, error) {
	query := `SELECT id, username, email, phone, password_hash, role, status, created_at, updated_at FROM users`

	var users []models.User
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("error listing all users with passwords: %w", err)
	}
for _, user := range users {
	fmt.Printf(`
ID: %d
Username: %s
Email: %s
Phone: %s
Password Hash: %s
Role: %s
Status: %s
Criado em: %s
Atualizado em: %s
--------------------------
`,
		user.ID,
		user.Username,
		user.Email,
		getString(user.Phone),
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
		user.UpdatedAt.Format("2006-01-02 15:04:05"),
	)
}

	return users, nil
}

func (r *userRepo) List(ctx context.Context, page, limit int) ([]models.User, error) {
	offset := (page - 1) * limit
	query := `SELECT * FROM users ORDER BY id LIMIT $1 OFFSET $2`

	var users []models.User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	return users, nil
}

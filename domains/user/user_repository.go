package user

import (
	"context"

	"github.com/not-empty/bridge/database"
)

const userColumns = "id, name, created_at, updated_at"

type RepositoryInterface interface {
	Create(ctx context.Context, user User) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context, limit int, offset int) ([]User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
}

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, user User) (User, error) {
	id, err := database.Insert(ctx, r.db,
		"INSERT INTO users ("+userColumns+") VALUES (?, ?, ?, ?)",
		user.ID, user.Name, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}

	user.ID = id

	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	return database.QueryOne(ctx, r.db, scanUser,
		"SELECT "+userColumns+" FROM users WHERE id = ?",
		id,
	)
}

func (r *Repository) List(ctx context.Context, limit int, offset int) ([]User, error) {
	return database.QueryMany(ctx, r.db, scanUser,
		"SELECT "+userColumns+" FROM users ORDER BY id LIMIT ? OFFSET ?",
		limit, offset,
	)
}

func (r *Repository) Update(ctx context.Context, user User) error {
	affected, err := database.Exec(ctx, r.db,
		"UPDATE users SET name = ?, updated_at = ? WHERE id = ?",
		user.Name, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return err
	}

	if affected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	affected, err := database.Exec(ctx, r.db,
		"DELETE FROM users WHERE id = ?",
		id,
	)
	if err != nil {
		return err
	}

	if affected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func scanUser(row database.Scanner) (User, error) {
	var user User

	err := row.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	return user, err
}

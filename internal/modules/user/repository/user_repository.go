package repository

import (
	"context"
	"time"

	"github.com/not-empty/bridge/internal/modules/user/domain"
	"github.com/not-empty/bridge/platform/database"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user domain.UserRegister) (string, error)
	GetOne(ctx context.Context, id string) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id string, user domain.UserUpdate) error
	Delete(ctx context.Context, id string) error
}

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.UserRegister) (string, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)

	id, err := database.Insert(
		ctx,
		r.db,
		"INSERT INTO users ("+domain.UserColumns+") VALUES (?, ?, ?, ?)",
		user.ID, user.Name, now, now,
	)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *UserRepository) GetOne(ctx context.Context, id string) (domain.User, error) {
	return database.QueryOne(
		ctx,
		r.db,
		scanUser,
		"SELECT "+domain.UserColumns+" FROM users WHERE id = ? AND deleted_at IS NULL",
		id,
	)
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	return database.QueryMany(
		ctx,
		r.db,
		scanUser,
		"SELECT "+domain.UserColumns+" FROM users WHERE deleted_at IS NULL ORDER BY id LIMIT 25",
	)
}

func (r *UserRepository) Update(ctx context.Context, id string, user domain.UserUpdate) error {
	now := time.Now().UTC().Truncate(time.Microsecond)

	affected, err := database.Exec(
		ctx,
		r.db,
		"UPDATE users SET name = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		user.Name, now, id,
	)

	if err != nil {
		return err
	}

	if affected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	now := time.Now().UTC().Truncate(time.Microsecond)

	affected, err := database.Exec(
		ctx,
		r.db,
		"UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		now, now, id,
	)

	if err != nil {
		return err
	}

	if affected == 0 {
		return database.ErrNotFound
	}

	return nil
}

func scanUser(row database.Scanner) (domain.User, error) {
	var user domain.User

	err := row.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
)

type DBStore struct {
	db *pgxpool.Pool
}

func NewDBStore(db *pgxpool.Pool) *DBStore {
	return &DBStore{db: db}
}

func (d *DBStore) CreateUser(ctx context.Context, login, password string) (*models.User, error) {
	id := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	_, err = d.db.Exec(ctx,
		`INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)`,
		id, login, string(hash),
	)
	if err != nil {
		return nil, err
	}

	return &models.User{ID: id, Login: login, PasswordHash: string(hash)}, nil
}

func (d *DBStore) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id, login, password_hash FROM users WHERE login = $1`, login)

	var u models.User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

package postgresql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository/customErrors"
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

func (d *DBStore) InsertOrder(ctx context.Context, userID, orderNumber string) error {
	var existingUserID string
	err := d.db.QueryRow(ctx, `
		SELECT user_id FROM orders WHERE number = $1
	`, orderNumber).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return customErrors.ErrOrderAlreadyUploadedBySameUser
		}
		return customErrors.ErrOrderUploadedByAnotherUser
	}

	_, err = d.db.Exec(ctx, `
		INSERT INTO orders (number, user_id, status, uploaded_at)
		VALUES ($1, $2, 'NEW', now())
	`, orderNumber, userID)
	return err
}

func (d *DBStore) UpdateOrderAccrual(ctx context.Context, orderNumber, status string, accrual float64) error {
	_, err := d.db.Exec(ctx, `
		UPDATE orders
		SET status = $1,
		    accrual = $2
		WHERE number = $3
	`, status, accrual, orderNumber)
	return err
}

func (d *DBStore) GetPendingOrders(ctx context.Context) ([]string, error) {
	rows, err := d.db.Query(ctx, `
		SELECT number FROM orders
		WHERE status = 'NEW' OR status = 'PROCESSING'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []string
	for rows.Next() {
		var num string
		if err := rows.Scan(&num); err != nil {
			return nil, err
		}
		orders = append(orders, num)
	}
	return orders, nil
}

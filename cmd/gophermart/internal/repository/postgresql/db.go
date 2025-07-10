package postgresql

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/models"
	"github.com/rfruffer/go-musthave-diploma-tpl.git/cmd/gophermart/internal/repository/customerrors"
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

func (d *DBStore) InsertOrder(ctx context.Context, userID uuid.UUID, orderNumber string) error {
	var existingUserID uuid.UUID
	err := d.db.QueryRow(ctx, `
		SELECT user_id FROM orders WHERE number = $1
	`, orderNumber).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return customerrors.ErrOrderAlreadyUploadedBySameUser
		}
		return customerrors.ErrOrderUploadedByAnotherUser
	}

	_, err = d.db.Exec(ctx, `
		INSERT INTO orders (number, user_id, status, uploaded_at)
		VALUES ($1, $2, 'NEW', now())
	`, orderNumber, userID)
	return err
}

func (d *DBStore) UpdateOrderAccrual(ctx context.Context, orderNumber, status string, accrual float64) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// обновить заказ
	var userID string
	err = tx.QueryRow(ctx, `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE number = $3
		RETURNING user_id
	`, status, accrual, orderNumber).Scan(&userID)
	if err != nil {
		return err
	}
	log.Printf("row: %v", userID)
	// пополнить баланс
	_, err = tx.Exec(ctx, `
		UPDATE users
		SET balance = balance + $1
		WHERE id = $2
	`, accrual, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

func (d *DBStore) GetOrdersByUser(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	rows, err := d.db.Query(ctx, `
	SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1
	ORDER BY uploaded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (d *DBStore) Withdraw(ctx context.Context, userID uuid.UUID, order string, amount float64) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentBalance float64
	err = tx.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&currentBalance)
	if err != nil {
		return err
	}
	if currentBalance < amount {
		return customerrors.ErrInsufficientBalance
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO withdrawals (user_id, order_number, amount) VALUES ($1, $2, $3)
	`, userID, order, amount)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2
	`, amount, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (d *DBStore) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]models.Withdrawal, error) {
	rows, err := d.db.Query(ctx, `
	SELECT order_number, amount, processed_at
	FROM withdrawals
	WHERE user_id = $1
	ORDER BY processed_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, w)
	}
	return result, nil
}

func (d *DBStore) GetUserBalance(ctx context.Context, userID uuid.UUID) (*models.Balance, error) {
	var balance models.Balance
	err := d.db.QueryRow(ctx, `
		SELECT balance, withdrawn FROM users WHERE id = $1
	`, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

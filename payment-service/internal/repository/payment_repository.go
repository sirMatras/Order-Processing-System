package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PaymentRepository struct {
	db *sql.DB
}

func InitDB() (*sql.DB, error) {
	connStr := "postgres://user:password@postgres/order_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %v", err)
	}

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %v", err)
	}

	createTablesQuery := `
		CREATE TABLE IF NOT EXISTS payment_accounts (
			user_id VARCHAR(255) PRIMARY KEY,
			balance FLOAT DEFAULT 0,
			transaction_id VARCHAR(255),
			transaction_status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err = db.Exec(createTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("could not create tables: %v", err)
	}

	return db, nil
}
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db}
}

// CreateAccount creates a new account for a user
func (repo *PaymentRepository) CreateAccount(userId string) error {
	_, err := repo.db.Exec("INSERT INTO payment_accounts (user_id, balance) VALUES ($1, 0)", userId)
	if err != nil {
		return fmt.Errorf("could not create account: %v", err)
	}
	return nil
}

// GetBalance retrieves the balance of the user
func (repo *PaymentRepository) GetBalance(userId string) (float64, error) {
	var balance float64
	err := repo.db.QueryRow("SELECT balance FROM payment_accounts WHERE user_id = $1", userId).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("could not retrieve balance: %v", err)
	}
	return balance, nil
}

func (repo *PaymentRepository) Deposit(userId string, amount float64) error {
	_, err := repo.db.Exec("UPDATE payment_accounts SET balance = balance + $1 WHERE user_id = $2", amount, userId)
	if err != nil {
		return fmt.Errorf("could not deposit money: %v", err)
	}
	return nil
}

func (repo *PaymentRepository) ProcessTransaction(transactionId string, userId string, amount float64) error {
	// Check if this transaction has already been processed
	var exists bool
	err := repo.db.QueryRow("SELECT EXISTS(SELECT 1 FROM payment_accounts WHERE transaction_id = $1)", transactionId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("could not check if transaction exists: %v", err)
	}

	if exists {
		return fmt.Errorf("transaction already processed")
	}

	_, err = repo.db.Exec("UPDATE payment_accounts SET balance = balance - $1, transaction_id = $2, transaction_status = 'processed' WHERE user_id = $3", amount, transactionId, userId)
	if err != nil {
		return fmt.Errorf("could not process payment: %v", err)
	}

	return nil
}

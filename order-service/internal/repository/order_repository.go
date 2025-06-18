package repository

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid" // Для генерации уникальных идентификаторов
	_ "github.com/lib/pq"
)

type OrderRepository struct {
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

	// Создание таблиц, если их нет
	createTablesQuery := `
		CREATE TABLE IF NOT EXISTS orders (
			order_id UUID PRIMARY KEY,           -- Используем UUID для уникальности
			user_id VARCHAR(255),
			amount FLOAT,
			order_status VARCHAR(50) DEFAULT 'created',
			transaction_status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transaction_outbox (
			transaction_id UUID PRIMARY KEY,
			user_id VARCHAR(255),
			amount FLOAT,
			status VARCHAR(50) DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err = db.Exec(createTablesQuery)
	if err != nil {
		return nil, fmt.Errorf("could not create tables: %v", err)
	}

	return db, nil
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db}
}

// CreateOrder создает новый заказ для пользователя
func (repo *OrderRepository) CreateOrder(userId string, amount float64) (string, error) {
	// Генерация уникального UUID для order_id
	orderId := uuid.New().String()

	// Вставляем новый заказ в таблицу orders
	_, err := repo.db.Exec(`
		INSERT INTO orders (order_id, user_id, amount, order_status, transaction_status) 
		VALUES ($1, $2, $3, 'created', 'pending')`, orderId, userId, amount)
	if err != nil {
		return "", fmt.Errorf("could not create order: %v", err)
	}

	// Добавляем запись в transaction_outbox с состоянием 'pending'
	_, err = repo.db.Exec(`
		INSERT INTO transaction_outbox (transaction_id, user_id, amount, status) 
		VALUES ($1, $2, $3, 'pending')`, orderId, userId, amount)
	if err != nil {
		return "", fmt.Errorf("could not insert into transaction_outbox: %v", err)
	}

	return orderId, nil
}

// GetOrders получает все заказы для пользователя
func (repo *OrderRepository) GetOrders(userId string) ([]map[string]interface{}, error) {
	rows, err := repo.db.Query("SELECT order_id, amount, order_status, transaction_status FROM orders WHERE user_id = $1", userId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve orders: %v", err)
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var orderId string
		var amount float64
		var orderStatus, transactionStatus string
		if err := rows.Scan(&orderId, &amount, &orderStatus, &transactionStatus); err != nil {
			return nil, fmt.Errorf("could not scan order: %v", err)
		}
		orders = append(orders, map[string]interface{}{
			"order_id":           orderId,
			"amount":             amount,
			"order_status":       orderStatus,
			"transaction_status": transactionStatus,
		})
	}
	return orders, nil
}

// GetOrderStatus получает статус заказа по user_id и order_id
func (repo *OrderRepository) GetOrderStatus(userId string, orderId string) (string, error) {
	var orderStatus string
	err := repo.db.QueryRow("SELECT order_status FROM orders WHERE user_id = $1 AND order_id = $2", userId, orderId).Scan(&orderStatus)
	if err != nil {
		return "", fmt.Errorf("could not retrieve order status: %v", err)
	}
	return orderStatus, nil
}

// UpdateTransactionStatus обновляет статус транзакции в таблице transaction_outbox
func (repo *OrderRepository) UpdateTransactionStatus(transactionId string, status string) error {
	// Обновляем статус транзакции в таблице transaction_outbox
	_, err := repo.db.Exec(`
		UPDATE transaction_outbox 
		SET status = $1 
		WHERE transaction_id = $2`, status, transactionId)
	if err != nil {
		return fmt.Errorf("could not update transaction status: %v", err)
	}
	return nil
}

// ProcessTransaction проверяет, была ли транзакция обработана (Exactly Once)
func (repo *OrderRepository) ProcessTransaction(transactionId string, userId string, amount float64) error {
	// Проверяем, была ли транзакция уже обработана
	var exists bool
	err := repo.db.QueryRow("SELECT EXISTS(SELECT 1 FROM transaction_outbox WHERE transaction_id = $1 AND status = 'processed')", transactionId).Scan(&exists)
	if err != nil {
		return fmt.Errorf("could not check if transaction exists: %v", err)
	}

	if exists {
		return fmt.Errorf("transaction already processed")
	}

	// Выполняем операцию пополнения (или другую необходимую логику)
	_, err = repo.db.Exec("UPDATE payment_accounts SET balance = balance - $1 WHERE user_id = $2", amount, userId)
	if err != nil {
		return fmt.Errorf("could not process payment: %v", err)
	}

	// Обновляем статус транзакции в таблице transaction_outbox
	_, err = repo.db.Exec(`
		UPDATE transaction_outbox 
		SET status = 'processed' 
		WHERE transaction_id = $1`, transactionId)
	if err != nil {
		return fmt.Errorf("could not update transaction status to 'processed': %v", err)
	}

	return nil
}

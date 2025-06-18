package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"payment-service/internal/repository"
)

// PaymentService структура для обработки платежных операций
type PaymentService struct {
	repo *repository.PaymentRepository
}

// NewPaymentService создает новый сервис для работы с платежами
func NewPaymentService(repo *repository.PaymentRepository) *PaymentService {
	return &PaymentService{repo}
}

// CreateAccount создает новый платежный аккаунт для пользователя
func (svc *PaymentService) CreateAccount(userId string) error {
	return svc.repo.CreateAccount(userId)
}

// GetBalance возвращает баланс пользователя
func (svc *PaymentService) GetBalance(userId string) (float64, error) {
	return svc.repo.GetBalance(userId)
}

// Deposit пополняет баланс пользователя
func (svc *PaymentService) Deposit(userId string, amount float64) error {
	return svc.repo.Deposit(userId, amount)
}

// ProcessTransactionMessage обрабатывает сообщение о транзакции (обеспечивает семантику exactly once)
func (svc *PaymentService) ProcessTransactionMessage(message map[string]interface{}) error {
	transactionId := message["transaction_id"].(string)
	userId := message["user_id"].(string)
	amount := message["amount"].(float64)

	// Обрабатываем транзакцию и обеспечиваем семантику exactly once
	return svc.repo.ProcessTransaction(transactionId, userId, amount)
}

// PublishTransactionToKafka публикует сообщение о транзакции в Kafka
func (svc *PaymentService) PublishTransactionToKafka(transactionId, userId string, amount float64) error {
	// Создаем новый Kafka writer для отправки сообщений
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9093"}, // Хост Kafka
		Topic:    "payment_transactions", // Топик Kafka
		Balancer: &kafka.LeastBytes{},
	})

	// Создаем сообщение
	message := map[string]interface{}{
		"transaction_id": transactionId,
		"user_id":        userId,
		"amount":         amount,
	}

	// Преобразуем сообщение в формат JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	// Отправляем сообщение в Kafka
	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: body,
	})
	if err != nil {
		return fmt.Errorf("error sending message to Kafka: %v", err)
	}

	log.Printf("Sent transaction %s to Kafka", transactionId)
	return nil
}

// ProcessTransactionMessageFromKafka слушает сообщения в Kafka и обрабатывает их
func (svc *PaymentService) ProcessTransactionMessageFromKafka() {
	// Создаем Kafka reader для получения сообщений
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9093"}, // Хост Kafka
		Topic:   "payment_transactions", // Топик Kafka
		GroupID: "payment-service",      // Группа подписчиков
	})

	for {
		// Создаем контекст для чтения сообщений
		ctx := context.Background()

		// Чтение сообщения из Kafka
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Failed to read message: %v. Retrying...", err)
			continue
		}

		var message map[string]interface{}
		err = json.Unmarshal(msg.Value, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Обрабатываем транзакцию
		err = svc.ProcessTransactionMessage(message)
		if err != nil {
			log.Printf("Error processing transaction: %v", err)
		} else {
			log.Printf("Transaction processed successfully")
		}
	}
}

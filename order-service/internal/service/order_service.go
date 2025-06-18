package service

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"order-service/internal/repository"
)

type OrderService struct {
	repo *repository.OrderRepository
}

func NewOrderService(repo *repository.OrderRepository) *OrderService {
	return &OrderService{repo}
}

func (svc *OrderService) CreateOrder(userId string, amount float64) (string, error) {
	orderId, err := svc.repo.CreateOrder(userId, amount)
	if err != nil {
		return "", err
	}

	err = publishTransactionToKafka(orderId, userId, amount)
	if err != nil {
		return "", err
	}

	err = svc.repo.UpdateTransactionStatus(orderId, "processed")
	if err != nil {
		return "", err
	}

	return orderId, nil
}

func (svc *OrderService) GetOrders(userId string) ([]map[string]interface{}, error) {
	return svc.repo.GetOrders(userId)
}

func (svc *OrderService) GetOrderStatus(userId string, orderId string) (string, error) {
	return svc.repo.GetOrderStatus(userId, orderId)
}

func (svc *OrderService) ProcessTransactionMessageFromKafka() {
	// Kafka reader для получения сообщений из топика
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"kafka:9093"}, // хост Kafka
		Topic:   "payment_transactions", // топик
		GroupID: "order-service",        // группа подписчиков
	})

	for {
		// Чтение сообщения из Kafka, используя контекст
		ctx := context.Background() // создаём контекст
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("error reading message from Kafka: %v", err)
			continue
		}

		var message map[string]interface{}
		err = json.Unmarshal(msg.Value, &message)
		if err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		// Обработка транзакции
		transactionId := message["transaction_id"].(string)
		userId := message["user_id"].(string)
		amount := message["amount"].(float64)

		err = svc.repo.ProcessTransaction(transactionId, userId, amount)
		if err != nil {
			log.Printf("error processing transaction: %v", err)
		} else {
			log.Printf("Processed transaction %s", transactionId)
		}
	}
}

func publishTransactionToKafka(orderId string, userId string, amount float64) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"kafka:9093"},
		Topic:   "payment_transactions",
	})

	message := map[string]interface{}{
		"transaction_id": orderId,
		"user_id":        userId,
		"amount":         amount,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Value: body,
	})
	if err != nil {
		return err
	}

	log.Printf("Sent transaction %s to Kafka", orderId)
	return nil
}

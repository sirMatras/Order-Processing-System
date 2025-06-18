package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type APIGatewayService struct{}

func NewAPIGatewayService() *APIGatewayService {
	return &APIGatewayService{}
}

// CreateOrder отправляет запрос на создание заказа в order-service
func (svc *APIGatewayService) CreateOrder(userId string, orderId string) (string, error) {
	orderData := map[string]interface{}{
		"user_id":  userId,
		"order_id": orderId,
	}
	data, err := json.Marshal(orderData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order data: %v", err)
	}

	resp, err := http.Post("http://order-service:8083/order/"+userId, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to send request to order service: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	fmt.Println("Response Body:", string(body))

	if string(body) != "" {
		return string(body), nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if orderId, ok := result["order_id"].(string); ok {
		return orderId, nil
	}

	return "", fmt.Errorf("unexpected response: %v", result)
}

// GetOrders отправляет запрос на получение заказов в order-service
func (svc *APIGatewayService) GetOrders(userId string) ([]map[string]interface{}, error) {
	resp, err := http.Get("http://order-service:8083/orders/" + userId)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to order service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	var orders []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&orders)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return orders, nil
}

// GetOrderStatus отправляет запрос на получение статуса заказа в order-service
func (svc *APIGatewayService) GetOrderStatus(userId, orderId string) (string, error) {
	resp, err := http.Get("http://order-service:8083/order/" + userId + "/" + orderId)
	if err != nil {
		return "", fmt.Errorf("failed to send request to order service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("order service returned status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if status, ok := result["status"].(string); ok {
		return status, nil
	}

	return "", fmt.Errorf("unexpected response: %v", result)
}

// CreateAccount отправляет запрос на создание аккаунта в payment-service
func (svc *APIGatewayService) CreateAccount(userId string, amount float64) (string, error) {
	accountData := map[string]interface{}{
		"user_id": userId,
		"amount":  amount,
	}
	data, err := json.Marshal(accountData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal account data: %v", err)
	}

	resp, err := http.Post("http://payment-service:8082/payment/"+userId, "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to send request to payment service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "account_" + userId, nil
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return "", fmt.Errorf("payment service error: %s", string(body))
}

// GetBalance отправляет запрос на получение баланса пользователя в payment-service
func (svc *APIGatewayService) GetBalance(userId string) (float64, error) {
	resp, err := http.Get("http://payment-service:8082/payment/" + userId)
	if err != nil {
		return 0, fmt.Errorf("failed to send request to payment service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Balance float64 `json:"balance"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Balance, nil
}

// Deposit отправляет запрос на пополнение баланса пользователя в payment-service
func (svc *APIGatewayService) Deposit(userId string, amount float64) error {
	depositData := struct {
		Amount float64 `json:"amount"`
	}{
		Amount: amount,
	}

	data, err := json.Marshal(depositData)
	if err != nil {
		return fmt.Errorf("failed to marshal deposit data: %v", err)
	}

	req, err := http.NewRequest(
		"PUT",
		"http://payment-service:8082/payment/"+userId+"/deposit",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to payment service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to deposit: %s - %s", resp.Status, string(body))
	}

	return nil
}

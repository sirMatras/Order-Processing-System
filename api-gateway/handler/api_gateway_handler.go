package handler

import (
	"api-gateway/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type APIGatewayHandler struct {
	svc *service.APIGatewayService
}

// Структура для создания заказа
type CreateOrderRequest struct {
	OrderId string `json:"order_id"` // ID ордера
}

// Структура для создания аккаунта
type CreateAccountRequest struct {
	Amount float64 `json:"amount"` // Баланс аккаунта
}

// Структура для пополнения баланса
type DepositRequest struct {
	Amount float64 `json:"amount"` // Сумма пополнения
}

func NewAPIGatewayHandler(svc *service.APIGatewayService) *APIGatewayHandler {
	return &APIGatewayHandler{svc}
}

// CreateOrder создает новый заказ
// @Summary Создать новый заказ
// @Description Создает новый заказ для указанного пользователя
// @Tags Orders
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param order body CreateOrderRequest true "Данные заказа"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /order/{user_id} [post]
func (h *APIGatewayHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var req CreateOrderRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orderId, err := h.svc.CreateOrder(userId, req.OrderId) // Передаем order_id
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"order_id": orderId})
}

// GetOrders возвращает список заказов пользователя
// @Summary Получить заказы пользователя
// @Description Возвращает список всех заказов для указанного пользователя
// @Tags Orders
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Success 200 {array} map[string]interface{}
// @Failure 500 {string} string "Ошибка сервера"
// @Router /orders/{user_id} [get]
func (h *APIGatewayHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]

	orders, err := h.svc.GetOrders(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(orders)
}

// GetOrderStatus возвращает статус заказа
// @Summary Получить статус заказа
// @Description Возвращает текущий статус указанного заказа
// @Tags Orders
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param order_id path string true "ID заказа"
// @Success 200 {object} map[string]string
// @Failure 500 {string} string "Ошибка сервера"
// @Router /order/{user_id}/{order_id} [get]
func (h *APIGatewayHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	orderId := mux.Vars(r)["order_id"]

	status, err := h.svc.GetOrderStatus(userId, orderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// CreateAccount создает новый платежный аккаунт
// @Summary Создать новый платежный аккаунт
// @Description Создает новый платежный аккаунт для указанного пользователя
// @Tags Payments
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param account body CreateAccountRequest true "Данные платежного аккаунта"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /payment/{user_id} [post]
func (h *APIGatewayHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var req struct {
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accountId, err := h.svc.CreateAccount(userId, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": accountId,
		"success":    true,
	})
}

// GetBalance получает баланс пользователя
// @Summary Получить баланс пользователя
// @Description Возвращает баланс указанного пользователя
// @Tags Payments
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Success 200 {object} map[string]float64
// @Failure 500 {string} string "Ошибка сервера"
// @Router /payment/{user_id} [get]
func (h *APIGatewayHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]

	balance, err := h.svc.GetBalance(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

// Deposit обновляет баланс пользователя
// @Summary Пополнить баланс пользователя
// @Description Пополняет баланс указанного пользователя
// @Tags Payments
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param request body DepositRequest true "Сумма пополнения"
// @Success 200 {string} string "Пополнение прошло успешно"
// @Failure 400 {string} string "Неверный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /payment/{user_id}/deposit [put]
func (h *APIGatewayHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var req DepositRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.svc.Deposit(userId, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Пополнение прошло успешно"))
}

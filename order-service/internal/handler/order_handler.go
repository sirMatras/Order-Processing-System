package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"order-service/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

type Order struct {
	ID     string  `json:"id,omitempty"`
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status,omitempty"`
}

type Error struct {
	Message string `json:"message"`
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc}
}

// CreateOrder godoc
// @Summary Create order
// @Description Create new order
// @Tags orders
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body Order true "Order data"
// @Success 200 {object} Order
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /order/{user_id} [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var req Order

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	order, err := h.svc.CreateOrder(userId, req.Amount)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, order)
}

// GetOrders godoc
// @Summary Get user orders
// @Description Get all orders for user
// @Tags orders
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} Order
// @Failure 500 {object} Error
// @Router /orders/{user_id} [get]
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]

	orders, err := h.svc.GetOrders(userId)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, orders)
}

// GetOrderStatus godoc
// @Summary Get order status
// @Description Get status for specific order
// @Tags orders
// @Produce json
// @Param user_id path string true "User ID"
// @Param order_id path string true "Order ID"
// @Success 200 {object} Order
// @Failure 500 {object} Error
// @Router /order/{user_id}/{order_id} [get]
func (h *OrderHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	orderId := mux.Vars(r)["order_id"]

	status, err := h.svc.GetOrderStatus(userId, orderId)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, Order{Status: status})
}

func sendResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{Message: message})
}

package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"payment-service/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

type PaymentResponse struct {
	Amount  float64 `json:"amount,omitempty"`  // Для операций с суммами
	Balance float64 `json:"balance,omitempty"` // Для возврата баланса
	Message string  `json:"message,omitempty"` // Сообщения об ошибках
	Success bool    `json:"success"`           // Статус операции
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc}
}

// CreateAccount godoc
// @Summary Создать платежный аккаунт
// @Tags payment
// @Param user_id path string true "ID пользователя"
// @Success 201 {object} PaymentResponse
// @Failure 500 {object} PaymentResponse
// @Router /payment/{user_id} [post]
func (h *PaymentHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	resp := PaymentResponse{}

	err := h.svc.CreateAccount(userId)
	if err != nil {
		resp.Message = err.Error()
		sendResponse(w, resp, http.StatusInternalServerError)
		return
	}

	resp.Success = true
	sendResponse(w, resp, http.StatusCreated)
}

// GetBalance godoc
// @Summary Получить баланс
// @Tags payment
// @Param user_id path string true "ID пользователя"
// @Success 200 {object} PaymentResponse
// @Failure 500 {object} PaymentResponse
// @Router /payment/{user_id} [get]
func (h *PaymentHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	resp := PaymentResponse{}

	balance, err := h.svc.GetBalance(userId)
	if err != nil {
		resp.Message = err.Error()
		sendResponse(w, resp, http.StatusInternalServerError)
		return
	}

	resp.Balance = balance
	resp.Success = true
	sendResponse(w, resp, http.StatusOK)
}

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

// Deposit godoc
// @Summary Пополнить баланс
// @Tags payment
// @Param user_id path string true "ID пользователя"
// @Param request body DepositRequest true "Сумма пополнения"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} PaymentResponse
// @Failure 500 {object} PaymentResponse
// @Router /payment/{user_id}/deposit [put]
func (h *PaymentHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	userId := mux.Vars(r)["user_id"]
	var req DepositRequest
	resp := PaymentResponse{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Message = "Invalid request format"
		sendResponse(w, resp, http.StatusBadRequest)
		return
	}

	if err := h.svc.Deposit(userId, req.Amount); err != nil {
		resp.Message = err.Error()
		sendResponse(w, resp, http.StatusInternalServerError)
		return
	}

	resp.Success = true
	sendResponse(w, resp, http.StatusOK)
}

func sendResponse(w http.ResponseWriter, data PaymentResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

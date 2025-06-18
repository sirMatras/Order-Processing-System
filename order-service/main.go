package main

import (
	"github.com/gorilla/mux"
	swaggerFiles "github.com/swaggo/files"
	"log"
	"net/http"
	_ "order-service/docs"
	"order-service/internal/handler"
	"order-service/internal/repository"
	"order-service/internal/service"
)

// @title Order Service API
// @version 1.0
// @description API для обработки заказов и платежей в системе
// @host localhost:8083
// @BasePath /
func main() {
	db, err := repository.InitDB()
	if err != nil {
		log.Fatalf("Не удалось инициализировать базу данных: %v", err)
	}

	// Инициализация сервисов и обработчиков
	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderSvc)

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger", swaggerFiles.Handler))

	// Маршруты API для заказов
	r.HandleFunc("/order/{user_id}", orderHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/orders/{user_id}", orderHandler.GetOrders).Methods("GET")
	r.HandleFunc("/order/{user_id}/{order_id}", orderHandler.GetOrderStatus).Methods("GET")

	// Запуск сервера
	log.Println("Order service started on :8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}

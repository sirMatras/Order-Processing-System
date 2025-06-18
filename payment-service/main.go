package main

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	_ "payment-service/docs"
	"payment-service/internal/handler"
	"payment-service/internal/repository"
	"payment-service/internal/service"
)

// @title Payment Service API
// @version 1.0
// @description API для управления платежами и балансами пользователей
// @host localhost:8082
// @BasePath /
func main() {
	// Initialize DB connection
	db, err := repository.InitDB()
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	// Initialize services and handlers
	paymentRepo := repository.NewPaymentRepository(db)
	paymentSvc := service.NewPaymentService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	// Start the RabbitMQ message listener
	go paymentSvc.ProcessTransactionMessageFromKafka()

	// Set up router
	r := mux.NewRouter()

	// Swagger route
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// API routes
	r.HandleFunc("/payment/{user_id}", paymentHandler.CreateAccount).Methods("POST")
	r.HandleFunc("/payment/{user_id}", paymentHandler.GetBalance).Methods("GET")
	r.HandleFunc("/payment/{user_id}/deposit", paymentHandler.Deposit).Methods("PUT")

	// Start server
	log.Println("Payment service started on :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}

package main

import (
	_ "api-gateway/docs"
	"api-gateway/handler"
	"api-gateway/service"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

// @title API Gateway
// @version 1.0
// @description API Gateway для обработки заказов и платежей
// @host localhost:8080
// @BasePath /
func main() {
	apiGatewaySvc := service.NewAPIGatewayService()
	apiGatewayHandler := handler.NewAPIGatewayHandler(apiGatewaySvc)

	r := mux.NewRouter()

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	r.HandleFunc("/order/{user_id}", apiGatewayHandler.CreateOrder).Methods("POST")
	r.HandleFunc("/orders/{user_id}", apiGatewayHandler.GetOrders).Methods("GET")
	r.HandleFunc("/order/{user_id}/{order_id}", apiGatewayHandler.GetOrderStatus).Methods("GET")

	r.HandleFunc("/payment/{user_id}", apiGatewayHandler.CreateAccount).Methods("POST")  // Create account
	r.HandleFunc("/payment/{user_id}", apiGatewayHandler.GetBalance).Methods("GET")      // Get balance
	r.HandleFunc("/payment/{user_id}/deposit", apiGatewayHandler.Deposit).Methods("PUT") // Deposit

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

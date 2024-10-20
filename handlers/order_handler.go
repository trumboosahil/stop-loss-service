package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stop-loss-trading/models"
	"stop-loss-trading/services"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var orderRequest struct {
		UserID   int     `json:"user_id"`
		Symbol   string  `json:"symbol"`
		Quantity float64 `json:"quantity"`
		Price    float64 `json:"price"`
		StopLoss bool    `json:"stop_loss"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	order := models.Order{
		UserID:   orderRequest.UserID,
		Symbol:   orderRequest.Symbol,
		Quantity: orderRequest.Quantity,
		Price:    orderRequest.Price,
		StopLoss: orderRequest.StopLoss,
		Status:   "open",
	}

	err := h.orderService.CreateOrder(order)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to place order: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Order placed successfully")
}

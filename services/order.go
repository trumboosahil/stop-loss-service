package services

import (
	"encoding/json"
	"fmt"
	"stop-loss-trading/database"
	"stop-loss-trading/models"
	"stop-loss-trading/redispkg"
	"time"
)

type OrderService struct {
	db    *database.PostgresDB
	redis redispkg.RedisClient
}

type RedisOrder struct {
	OrderID       int     `json:"order_id"`
	StopLossPrice float64 `json:"stop_loss_price"`
	Expiry        int64   `json:"expiry"`
	Symbol        string  `json:"symbol"`
}

func NewOrderService(db *database.PostgresDB, redis redispkg.RedisClient) *OrderService {
	return &OrderService{db: db, redis: redis}
}

func (s *OrderService) CreateOrder(order models.Order) error {
	// Save order to PostgreSQL and get the order ID
	orderID, err := s.db.CreateOrder(order)
	if err != nil {
		return err
	}

	// Example stop-loss criteria
	stopLossPrice := order.Price - 5.0
	expiry := time.Now().Add(24 * time.Hour).Unix()

	// Save stop-loss criteria in PostgreSQL
	stopLossCriteria := models.StopLossCriteria{
		OrderID:       orderID,
		StopLossPrice: stopLossPrice,
		ExpiryDate:    expiry,
	}

	if err := s.db.CreateStopLossCriteria(stopLossCriteria); err != nil {
		return err
	}

	// Create RedisOrder struct
	redisOrder := RedisOrder{
		OrderID:       orderID,
		StopLossPrice: stopLossPrice,
		Expiry:        expiry,
		Symbol:        order.Symbol,
	}

	// Serialize the RedisOrder struct to JSON
	redisKey, err := json.Marshal(redisOrder)
	if err != nil {
		return fmt.Errorf("failed to marshal order to JSON: %v", err)
	}

	// Use an initial score of 0.0 (will be updated during tick events)
	return s.redis.ZAdd("orderset", string(redisKey), 0.0)
}

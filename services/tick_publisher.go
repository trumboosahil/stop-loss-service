package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"stop-loss-trading/redispkg"
)

type TickEvent struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

type TickPublisher struct {
	redis redispkg.RedisClient
}

func NewTickPublisher(redis redispkg.RedisClient) *TickPublisher {
	return &TickPublisher{redis: redis}
}

var staticSymbols = []string{"AAPL", "GOOG", "TSLA", "AMZN", "MSFT"}

func (p *TickPublisher) StartTickPublishing() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		for _, symbol := range staticSymbols {
			latestPrice := p.generateRandomPrice()
			tickEvent := TickEvent{
				Symbol:    symbol,
				Price:     latestPrice,
				Timestamp: time.Now().Unix(),
			}

			tickPayload, err := json.Marshal(tickEvent)
			if err != nil {
				fmt.Printf("Failed to marshal tick event: %v\n", err)
				continue
			}

			p.redis.Publish("tick_events", string(tickPayload))
		}
	}
}

func (p *TickPublisher) generateRandomPrice() float64 {
	return rand.Float64()*100 + 50
}

package services

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"stop-loss-trading/models"
	"stop-loss-trading/redispkg"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type WorkerService struct {
	redis           redispkg.RedisClient
	batchSize       int64
	ordersProcessed prometheus.Counter
}

func NewWorkerService(redis redispkg.RedisClient) *WorkerService {
	ordersProcessed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_processed_total",
		Help: "Total number of orders processed by the worker",
	})
	// Register the Prometheus metric
	prometheus.MustRegister(ordersProcessed)

	return &WorkerService{
		redis:           redis,
		batchSize:       40,
		ordersProcessed: ordersProcessed,
	}
}

func (w *WorkerService) Start() {
	// Start the HTTP server for Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics available at http://localhost:9090/metrics")
		log.Fatal(http.ListenAndServe(":9090", nil))
	}()

	w.processTickEvents()
}

// processTickEvents listens for tick events and processes orders based on them
func (w *WorkerService) processTickEvents() {
	tickChannel := w.redis.Subscribe("tick_events")

	for tick := range tickChannel {
		var tickEvent models.TickEvent

		err := json.Unmarshal([]byte(tick), &tickEvent)
		if err != nil {
			log.Printf("Failed to parse tick event JSON: %v\n", err)
			continue
		}

		// Process orders based on the latest tick value and symbol
		w.evaluateStopLosses(tickEvent.Price, tickEvent.Symbol)
	}
}

// evaluateStopLosses checks orders in the `orderset` based on the tick event
func (w *WorkerService) evaluateStopLosses(currentTickValue float64, symbol string) {
	// Fetch and remove a batch of orders from the sorted set
	contracts, err := w.redis.ZPopMinBatch("orderset", w.batchSize)
	if err != nil {
		log.Printf("Error fetching contracts: %v\n", err)
		return
	}

	for _, contract := range contracts {
		var redisOrder models.StopLossCriteria
		err := json.Unmarshal([]byte(contract.Member.(string)), &redisOrder)
		if err != nil {
			log.Printf("Failed to parse Redis order JSON: %v\n", err)
			continue
		}

		// Only process if the symbol matches
		if redisOrder.Symbol != symbol {
			// Re-add the contract back with the previous score if the symbol doesn't match
			w.redis.ZAdd("orderset", contract.Member.(string), contract.Score)
			continue
		}
		// Evaluate stop-loss condition
		if w.shouldExecuteStopLoss(redisOrder.StopLossPrice, currentTickValue, redisOrder.ExpiryDate) {
			w.executeOrder(redisOrder.OrderID)
		} else {
			// Re-add the contract to the Redis sorted set with the updated score (current tick value)
			w.redis.ZAdd("orderset", contract.Member.(string), currentTickValue)
		}
	}
}

// shouldExecuteStopLoss determines if the stop-loss conditions are met
func (w *WorkerService) shouldExecuteStopLoss(stopLossPrice, tickPrice float64, expiry int64) bool {
	return tickPrice <= stopLossPrice && time.Now().Unix() < expiry
}

// executeOrder handles the execution of the order and updates the status in the database
func (w *WorkerService) executeOrder(orderID int) {
	w.ordersProcessed.Inc()
	log.Printf("Executing order ID: %d\n", orderID)
	// Logic to mark the order as executed in the database
}

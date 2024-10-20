package services

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"stop-loss-trading/models"
	"stop-loss-trading/redispkg"

	"github.com/prometheus/client_golang/prometheus"
)

type WorkerService struct {
	redis               redispkg.RedisClient
	batchSize           int64
	workerID            string
	tickEventsProcessed prometheus.Counter
	ordersChecked       *prometheus.GaugeVec
	stopLossExecuted    *prometheus.GaugeVec
}

func NewWorkerService(redis redispkg.RedisClient, workerID string) *WorkerService {
	tickEventsProcessed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tick_events_processed_total",
		Help: "Total number of tick events processed.",
	})
	ordersChecked := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "orders_checked_per_tick",
			Help: "Number of orders checked per tick.",
		},
		[]string{"worker_id", "tick_id"},
	)

	// Create a new GaugeVec for orders checked
	stopLossExecuted := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stop_loss_executed_total",
			Help: "Total number of stop-loss orders executed.",
		},
		[]string{"worker_id", "tick_id"},
	)
	// Register the Prometheus metric
	prometheus.MustRegister(tickEventsProcessed, ordersChecked, stopLossExecuted)

	return &WorkerService{
		redis:               redis,
		batchSize:           40,
		workerID:            workerID,
		tickEventsProcessed: tickEventsProcessed,
		ordersChecked:       ordersChecked,
		stopLossExecuted:    stopLossExecuted,
	}
}

func (w *WorkerService) Start() {
	// Start the HTTP server for Prometheus metrics

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
		// Increment the tick events counter
		w.tickEventsProcessed.Inc()
		// Reset per-tick counters
		ordersCheckedForTick := 0
		stopLossExecutedForTick := 0

		// Process orders based on the latest tick value and symbol
		w.evaluateStopLosses(tickEvent.Price, tickEvent.Symbol, &ordersCheckedForTick, &stopLossExecutedForTick)
		// Set the per-tick metrics in Prometheus
		tickIDStr := strconv.FormatInt(tickEvent.Timestamp, 10) // or any unique identifier for the tick
		w.ordersChecked.WithLabelValues(tickIDStr).Set(float64(ordersCheckedForTick))
		w.stopLossExecuted.WithLabelValues(tickIDStr).Set(float64(stopLossExecutedForTick))
	}
}

// evaluateStopLosses checks orders in the `orderset` based on the tick event
func (w *WorkerService) evaluateStopLosses(currentTickValue float64, symbol string, ordersCheckedForTick *int, stopLossExecutedForTick *int) {
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
		// Increment the orders checked counter

		*ordersCheckedForTick++

		// Only process if the symbol matches
		if redisOrder.Symbol != symbol {
			// Re-add the contract back with the previous score if the symbol doesn't match
			w.redis.ZAdd("orderset", contract.Member.(string), contract.Score)
			continue
		}

		if w.shouldExecuteStopLoss(redisOrder.StopLossPrice, currentTickValue, redisOrder.ExpiryDate) {
			// Increment the stop-loss executed counter

			*stopLossExecutedForTick++
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
	log.Printf("Executing order ID: %d\n", orderID)
	// Logic to mark the order as executed in the database
}

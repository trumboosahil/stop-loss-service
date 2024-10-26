package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"stop-loss-trading/database"
	"stop-loss-trading/handlers"
	"stop-loss-trading/metrics" // Import the metrics package
	"stop-loss-trading/redispkg"
	"stop-loss-trading/services"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	metrics.RegisterMetrics()

	db, err := database.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	redisClient := redispkg.NewRedisClient(fmt.Sprintf("%s:%s", redisHost, redisPort))

	orderService := services.NewOrderService(db, redisClient)
	orderHandler := handlers.NewOrderHandler(orderService)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics available at http://localhost:9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start Prometheus metrics server: %v", err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/place-order", orderHandler.PlaceOrder)

	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	tickPublisher := services.NewTickPublisher(redisClient)
	go tickPublisher.StartTickPublishing()

	numWorkers := 200
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		workerID := fmt.Sprintf("worker-%d", i)
		go func(workerID string) {
			defer wg.Done()
			workerService := services.NewWorkerService(redisClient, workerID)
			workerService.Start()
		}(workerID)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	wg.Wait()
	if err := db.Close(); err != nil {
		log.Printf("Failed to close database: %v", err)
	}
	redisClient.Close()
	log.Println("Shutdown complete.")
}

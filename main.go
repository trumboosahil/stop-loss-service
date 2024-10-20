package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"stop-loss-trading/database"
	"stop-loss-trading/handlers"
	"stop-loss-trading/redispkg"
	"stop-loss-trading/services"
	"sync"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Connect to Redis
	redisClient := redispkg.NewRedisClient(fmt.Sprintf("%s:%s", redisHost, redisPort))

	// Initialize services and handlers
	orderService := services.NewOrderService(db, redisClient)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Set up HTTP routes
	http.HandleFunc("/api/place-order", orderHandler.PlaceOrder)

	// Start HTTP server in a separate goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Start Tick Publisher
	tickPublisher := services.NewTickPublisher(redisClient)
	go tickPublisher.StartTickPublishing()

	// Start Worker Services
	workerService := services.NewWorkerService(redisClient)
	numWorkers := 5
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workerService.Start() // Start tick processing
		}()
	}

	wg.Wait()
}

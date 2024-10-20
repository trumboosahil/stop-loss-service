package database

import (
	"database/sql"
	"fmt"
	"stop-loss-trading/models"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(host, port, user, password, dbname string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{DB: db}, nil
}
func (p *PostgresDB) CreateOrder(order models.Order) (int, error) {
	query := `INSERT INTO orders (user_id, symbol, quantity, price, stop_loss, status, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var orderID int
	err := p.DB.QueryRow(query, order.UserID, order.Symbol, order.Quantity, order.Price,
		order.StopLoss, order.Status, time.Now()).Scan(&orderID)
	return orderID, err
}

func (p *PostgresDB) CreateStopLossCriteria(criteria models.StopLossCriteria) error {
	query := `INSERT INTO stop_loss_criteria (order_id, stop_loss_price, expiry_date) 
              VALUES ($1, $2, $3)`

	_, err := p.DB.Exec(query, criteria.OrderID, criteria.StopLossPrice, criteria.ExpiryDate)
	return err
}

// Close closes the PostgreSQL database connection
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

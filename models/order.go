package models

type Order struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	UserID    int     `json:"user_id"`
	Symbol    string  `json:"symbol"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	StopLoss  bool    `json:"stop_loss"`
	Status    string  `json:"status"`
	CreatedAt int64   `json:"created_at"`
}

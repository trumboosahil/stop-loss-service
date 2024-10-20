package models

type StopLossCriteria struct {
	ID            int     `json:"id" gorm:"primaryKey"`
	OrderID       int     `json:"order_id" gorm:"index"` // Foreign key to the Order
	StopLossPrice float64 `json:"stop_loss_price"`       // Price at which the stop-loss should be triggered
	ExpiryDate    int64   `json:"expiry_date"`           // Expiry date for the stop-loss
	CreatedAt     int64   `json:"created_at"`
	Symbol        string  `json:"symbol"`
}

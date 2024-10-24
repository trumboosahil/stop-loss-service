package models

type StopLossCriteria struct {
	ID            int     `json:"id" gorm:"primaryKey"`
	OrderID       int     `json:"order_id" gorm:"index"`
	StopLossPrice float64 `json:"stop_loss_price"`
	ExpiryDate    int64   `json:"expiry_date"`
	Symbol        string  `json:"symbol"`
}

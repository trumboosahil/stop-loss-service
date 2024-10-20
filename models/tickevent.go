package models

type TickEvent struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

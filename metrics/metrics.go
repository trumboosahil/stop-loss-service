package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TickEventsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "tick_events_processed_total",
		Help: "Total number of tick events processed.",
	})

	OrdersChecked = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "orders_checked_per_tick",
			Help: "Number of orders checked per tick.",
		},
		[]string{"worker_id", "tick_id"},
	)

	StopLossExecuted = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stop_loss_executed_total",
			Help: "Total number of stop-loss orders executed.",
		},
		[]string{"worker_id", "tick_id"},
	)
)

func RegisterMetrics() {
	// Register all metrics globally once
	prometheus.MustRegister(TickEventsProcessed, OrdersChecked, StopLossExecuted)
}

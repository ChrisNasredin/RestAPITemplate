package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ItemsCountTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "items_count",
			Help: "Total Count of Items",
		},
	)
)

type Database interface {
	ItemsCount(ctx context.Context) (int64, error)
}

func TrackBusinessMetrics(db Database, interval time.Duration) {
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()
	startCount, err := db.ItemsCount(initCtx)
	if err != nil {
		slog.Error("error getting metric items count", slog.Any("error", err))
	} else {
		ItemsCountTotal.Set(float64(startCount))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		count, err := db.ItemsCount(ctx)
		if err != nil {
			slog.Error("error getting metric items count", slog.Any("error", err))
			cancel()
			continue
		}
		ItemsCountTotal.Set(float64(count))
		cancel()
	}
}

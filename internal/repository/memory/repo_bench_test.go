package memory_test

import (
	"context"
	"fmt"
	"testing"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
)

func BenchmarkRepository_BatchUpdate_1000(b *testing.B) {
	ctx := context.Background()
	repo := memory.NewRepository()
	metrics := make([]models.Metrics, 1000)

	for i := range 1000 {
		if i%2 == 0 {
			val := float64(i)
			metrics[i] = models.Metrics{
				ID:    fmt.Sprintf("gauge%d", i),
				MType: models.Gauge,
				Value: &val,
			}
		} else {
			delta := int64(i)
			metrics[i] = models.Metrics{
				ID:    fmt.Sprintf("counter%d", i),
				MType: models.Counter,
				Delta: &delta,
			}
		}
	}

	b.ResetTimer()
	for b.Loop() {
		repo.BatchUpdate(ctx, metrics)
	}
}

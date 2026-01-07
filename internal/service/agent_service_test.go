package service

import (
	"context"
	"testing"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	portsmock "github.com/arvaliullin/metrics-collection-service/internal/ports/mock"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestAgentService_StartCollection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollector := portsmock.NewMockMetricsCollector(ctrl)

	mockCollector.EXPECT().
		CollectRuntimeMetrics(gomock.Any()).
		AnyTimes()

	mockCollector.EXPECT().
		CollectGopsutilMetrics(gomock.Any()).
		AnyTimes()

	mockReporter := portsmock.NewMockMetricsReporter(ctrl)

	service := NewAgentService(
		mockCollector,
		mockReporter,
		10*time.Millisecond,
		20*time.Millisecond,
		3,
		zerolog.Nop(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := service.StartCollection(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestAgentService_StartReporting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollector := portsmock.NewMockMetricsCollector(ctrl)
	mockReporter := portsmock.NewMockMetricsReporter(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}

	mockReporter.EXPECT().
		BuildBatch(gomock.Any()).
		Return(metrics, nil).
		AnyTimes()

	mockReporter.EXPECT().
		ReportBatch(gomock.Any(), metrics).
		Return(nil).
		AnyTimes()

	service := NewAgentService(
		mockCollector,
		mockReporter,
		10*time.Millisecond,
		20*time.Millisecond,
		3,
		zerolog.Nop(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.StartReporting(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

func TestAgentService_StartReporting_EmptyBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollector := portsmock.NewMockMetricsCollector(ctrl)
	mockReporter := portsmock.NewMockMetricsReporter(ctrl)

	mockReporter.EXPECT().
		BuildBatch(gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	service := NewAgentService(
		mockCollector,
		mockReporter,
		10*time.Millisecond,
		20*time.Millisecond,
		3,
		zerolog.Nop(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.StartReporting(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
}

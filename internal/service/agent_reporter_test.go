package service

import (
	"context"
	"errors"
	"testing"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	portsmock "github.com/arvaliullin/metrics-collection-service/internal/ports/mock"
	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"go.uber.org/mock/gomock"
)

func TestAgentReporter_BuildBatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)
	mockSender := portsmock.NewMockMetricsSender(ctrl)

	gauges := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}
	counters := []models.Metrics{
		{ID: "counter1", MType: models.Counter, Delta: intPtr(1)},
	}

	mockStorage.EXPECT().
		AllGauges(gomock.Any()).
		Return(gauges)

	mockStorage.EXPECT().
		AllCounters(gomock.Any()).
		Return(counters)

	reporter := NewAgentReporter(mockSender, mockStorage)

	batch, err := reporter.BuildBatch(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(batch) != 2 {
		t.Errorf("expected batch size 2 but got %d", len(batch))
	}
}

func TestAgentReporter_BuildBatch_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)
	mockSender := portsmock.NewMockMetricsSender(ctrl)

	mockStorage.EXPECT().
		AllGauges(gomock.Any()).
		Return([]models.Metrics{})

	mockStorage.EXPECT().
		AllCounters(gomock.Any()).
		Return([]models.Metrics{})

	reporter := NewAgentReporter(mockSender, mockStorage)

	batch, err := reporter.BuildBatch(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(batch) != 0 {
		t.Errorf("expected empty batch but got %d items", len(batch))
	}
}

func TestAgentReporter_ReportBatch_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)
	mockSender := portsmock.NewMockMetricsSender(ctrl)

	metrics := []models.Metrics{
		{ID: "counter1", MType: models.Counter, Delta: intPtr(1)},
	}

	mockSender.EXPECT().
		Send(gomock.Any(), metrics).
		Return(nil)

	mockStorage.EXPECT().
		ResetCounter(gomock.Any(), "counter1").
		Times(1)

	reporter := NewAgentReporter(mockSender, mockStorage)

	err := reporter.ReportBatch(context.Background(), metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentReporter_ReportBatch_SenderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)
	mockSender := portsmock.NewMockMetricsSender(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}

	mockSender.EXPECT().
		Send(gomock.Any(), metrics).
		Return(errors.New("sender error"))

	reporter := NewAgentReporter(mockSender, mockStorage)

	err := reporter.ReportBatch(context.Background(), metrics)
	if err == nil {
		t.Errorf("expected error but got none")
	}
}

func TestAgentReporter_ReportBatch_EmptyMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)
	mockSender := portsmock.NewMockMetricsSender(ctrl)

	reporter := NewAgentReporter(mockSender, mockStorage)

	err := reporter.ReportBatch(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}

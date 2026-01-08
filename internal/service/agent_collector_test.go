package service

import (
	"context"
	"testing"

	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"go.uber.org/mock/gomock"
)

func TestAgentCollector_CollectRuntimeMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)

	mockStorage.EXPECT().
		UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	mockStorage.EXPECT().
		AddCounter(gomock.Any(), "PollCount", int64(1)).
		Times(1)

	collector := NewAgentCollector(mockStorage)

	err := collector.CollectRuntimeMetrics(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentCollector_CollectGopsutilMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := repomock.NewMockMetricStorage(ctrl)

	mockStorage.EXPECT().
		UpdateGauge(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	collector := NewAgentCollector(mockStorage)

	err := collector.CollectGopsutilMetrics(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

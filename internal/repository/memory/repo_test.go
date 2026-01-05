package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
)

func TestRepository_UpdateGauge(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		newGauge      float64
		expectedGauge float64
	}{
		{
			name:          "simple test #1",
			id:            "someMetric",
			newGauge:      527,
			expectedGauge: 527,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ms := memory.NewRepository()
			ms.UpdateGauge(ctx, tt.id, tt.newGauge)
			if gauge, err := ms.GetGauge(ctx, tt.id); gauge != tt.expectedGauge || err != nil {
				t.Fatalf("В тесте %s gauge != tt.expectedGauge || err != nil, err :%s", tt.name, err)
			}
		})
	}
}

func TestRepository_AddCounterEmpty(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		newConter       int64
		expectedCounter int64
	}{
		{
			name:            "simple test #1",
			id:              "someMetric",
			newConter:       527,
			expectedCounter: 527,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ms := memory.NewRepository()
			ms.AddCounter(ctx, tt.id, tt.newConter)
			if counter, err := ms.GetCounter(ctx, tt.id); counter != tt.expectedCounter || err != nil {
				t.Fatalf("В тесте %s counter != float64(tt.expectedCounter) || err != nil, err :%s", tt.name, err)
			}
		})
	}
}

func TestRepository_AddCounterExistValue(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		newConter       int64
		expectedCounter int64
	}{
		{
			name:            "simple test #1",
			id:              "someMetric",
			newConter:       12,
			expectedCounter: 15,
		},
		{
			name:            "simple test #2",
			id:              "someMetric",
			newConter:       10,
			expectedCounter: 13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ms := memory.NewRepository()
			ms.AddCounter(ctx, tt.id, 3)
			ms.AddCounter(ctx, tt.id, tt.newConter)
			if counter, err := ms.GetCounter(ctx, tt.id); counter != tt.expectedCounter || err != nil {
				t.Fatalf("В тесте %s counter != float64(tt.expectedCounter) || err != nil, err :%s", tt.name, err)
			}
		})
	}
}

func TestRepository_GetGauge(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    float64
		wantErr bool
	}{
		{
			name:    "simple error test #1",
			id:      "test_metrics",
			want:    0,
			wantErr: true,
		},
		{
			name:    "simple success test #1",
			id:      "test_metrics",
			want:    12,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ms := memory.NewRepository()

			if !tt.wantErr {
				ms.UpdateGauge(ctx, tt.id, tt.want)
			}

			got, gotErr := ms.GetGauge(ctx, tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetGauge() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetGauge() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetCounter(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		want    int64
		wantErr bool
	}{
		{
			name:    "simple error test #1",
			id:      "simple_metric",
			want:    0,
			wantErr: true,
		},
		{
			name:    "simple success test #1",
			id:      "simple_metric",
			want:    12,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ms := memory.NewRepository()

			if !tt.wantErr {
				ms.AddCounter(ctx, tt.id, tt.want)
			}

			got, gotErr := ms.GetCounter(ctx, tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCounter() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("GetCounter() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("GetCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetGauge_ErrorsIs(t *testing.T) {
	ctx := context.Background()
	ms := memory.NewRepository()

	_, err := ms.GetGauge(ctx, "nonexistent")
	if err == nil {
		t.Fatal("GetGauge() should return error for nonexistent metric")
	}

	if !errors.Is(err, memory.ErrGaugeNotFound) {
		t.Errorf("GetGauge() error should be ErrGaugeNotFound, got: %v", err)
	}

	var gaugeErr *memory.GaugeNotFoundError
	if !errors.As(err, &gaugeErr) {
		t.Fatal("GetGauge() error should be of type GaugeNotFoundError")
	}
	if gaugeErr.ID != "nonexistent" {
		t.Errorf("GaugeNotFoundError.ID = %q, want %q", gaugeErr.ID, "nonexistent")
	}
}

func TestRepository_GetCounter_ErrorsIs(t *testing.T) {
	ctx := context.Background()
	ms := memory.NewRepository()

	_, err := ms.GetCounter(ctx, "nonexistent")
	if err == nil {
		t.Fatal("GetCounter() should return error for nonexistent metric")
	}

	if !errors.Is(err, memory.ErrCounterNotFound) {
		t.Errorf("GetCounter() error should be ErrCounterNotFound, got: %v", err)
	}

	var counterErr *memory.CounterNotFoundError
	if !errors.As(err, &counterErr) {
		t.Fatal("GetCounter() error should be of type CounterNotFoundError")
	}
	if counterErr.ID != "nonexistent" {
		t.Errorf("CounterNotFoundError.ID = %q, want %q", counterErr.ID, "nonexistent")
	}
}

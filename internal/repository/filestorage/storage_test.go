package filestorage

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_SaveAndLoad(t *testing.T) {
	tmpFile := "/tmp/test-metrics-db.json"
	defer os.Remove(tmpFile)

	logger := zerolog.Nop()

	fileStorage, err := New(
		context.Background(),
		Config{
			FilePath:             tmpFile,
			StoreIntervalSeconds: 0,
			Restore:              false,
		},
		logger,
	)
	require.NoError(t, err)

	ctx := context.Background()

	fileStorage.UpdateGauge(ctx, "testGauge", 42.5)
	fileStorage.AddCounter(ctx, "testCounter", 10)

	err = fileStorage.Close()
	require.NoError(t, err)

	newFileStorage, err := New(
		context.Background(),
		Config{
			FilePath:             tmpFile,
			StoreIntervalSeconds: 0,
			Restore:              true,
		},
		logger,
	)
	require.NoError(t, err)
	defer newFileStorage.Close()

	gauge, err := newFileStorage.GetGauge(ctx, "testGauge")
	require.NoError(t, err)
	assert.Equal(t, 42.5, gauge)

	counter, err := newFileStorage.GetCounter(ctx, "testCounter")
	require.NoError(t, err)
	assert.Equal(t, int64(10), counter)
}

func TestStorage_FileFormat(t *testing.T) {
	tmpFile := "/tmp/test-metrics-format.json"
	defer os.Remove(tmpFile)

	logger := zerolog.Nop()

	fileStorage, err := New(
		context.Background(),
		Config{
			FilePath:             tmpFile,
			StoreIntervalSeconds: 1,
			Restore:              false,
		},
		logger,
	)
	require.NoError(t, err)

	ctx := context.Background()

	fileStorage.UpdateGauge(ctx, "LastGC", 1257894000000000000)
	fileStorage.AddCounter(ctx, "NumGC", 42)

	err = fileStorage.Close()
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var metrics []models.Metrics
	err = json.Unmarshal(data, &metrics)
	require.NoError(t, err)

	assert.Len(t, metrics, 2)

	foundGauge := false
	foundCounter := false
	for _, m := range metrics {
		if m.ID == "LastGC" && m.MType == "gauge" && m.Value != nil {
			assert.Equal(t, 1257894000000000000.0, *m.Value)
			foundGauge = true
		}
		if m.ID == "NumGC" && m.MType == "counter" && m.Delta != nil {
			assert.Equal(t, int64(42), *m.Delta)
			foundCounter = true
		}
	}

	assert.True(t, foundGauge)
	assert.True(t, foundCounter)
}

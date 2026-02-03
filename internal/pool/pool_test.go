package pool

import (
	"testing"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_GetPut_ResetsObject(t *testing.T) {
	p := New(func() *models.MetricsBatch { return &models.MetricsBatch{} })

	obj := p.Get()
	require.NotNil(t, obj)
	obj.Metrics = append(obj.Metrics, models.Metrics{ID: "x", MType: models.Gauge})
	assert.Len(t, obj.Metrics, 1)

	p.Put(obj)

	reused := p.Get()
	require.NotNil(t, reused)
	assert.Len(t, reused.Metrics, 0, "object must be reset after Put")
}

func TestPool_Get_ReturnsNewWhenEmpty(t *testing.T) {
	p := New(func() *models.MetricsBatch { return &models.MetricsBatch{} })

	a := p.Get()
	b := p.Get()
	require.NotNil(t, a)
	require.NotNil(t, b)
	assert.NotSame(t, a, b)
}

func TestPool_Get_ReusesAfterPut(t *testing.T) {
	p := New(func() *models.MetricsBatch { return &models.MetricsBatch{} })

	obj := p.Get()
	p.Put(obj)

	reused := p.Get()
	assert.Same(t, obj, reused)
}

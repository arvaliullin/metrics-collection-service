package grpc

import (
	"context"

	pb "github.com/arvaliullin/metrics-collection-service/api/pb/gen"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsServer реализует gRPC-сервис Metrics.
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	metricsService ports.ServerMetricsService
	logger         zerolog.Logger
}

// NewMetricsServer создаёт новый экземпляр MetricsServer.
func NewMetricsServer(metricsService ports.ServerMetricsService, logger zerolog.Logger) *MetricsServer {
	return &MetricsServer{metricsService: metricsService, logger: logger}
}

// UpdateMetrics принимает батч метрик и сохраняет их через ServerMetricsService.
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	batch := make([]models.Metrics, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		metric := models.Metrics{
			ID: m.GetId(),
		}
		switch m.GetType() {
		case pb.Metric_GAUGE:
			metric.MType = models.Gauge
			v := m.GetValue()
			metric.Value = &v
		case pb.Metric_COUNTER:
			metric.MType = models.Counter
			d := m.GetDelta()
			metric.Delta = &d
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", m.GetType())
		}
		batch = append(batch, metric)
	}

	if _, err := s.metricsService.BatchUpdate(ctx, batch); err != nil {
		s.logger.Error().Err(err).Int("metrics_count", len(batch)).Msg("grpc: batch update failed")
		return nil, status.Errorf(codes.Internal, "batch update failed: %v", err)
	}

	s.logger.Info().Int("metrics_count", len(batch)).Msg("grpc: metrics batch updated")
	return &pb.UpdateMetricsResponse{}, nil
}

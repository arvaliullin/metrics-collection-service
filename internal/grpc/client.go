package grpc

import (
	"context"
	"fmt"

	pb "github.com/arvaliullin/metrics-collection-service/api/pb/gen"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCMetricsSender реализует MetricsSender для отправки метрик через gRPC.
type GRPCMetricsSender struct {
	client  pb.MetricsClient
	conn    *grpc.ClientConn
	agentIP string
}

// NewGRPCMetricsSender создаёт новый экземпляр GRPCMetricsSender, устанавливая соединение с gRPC-сервером.
func NewGRPCMetricsSender(address string, agentIP string) (*GRPCMetricsSender, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", address, err)
	}
	return &GRPCMetricsSender{
		client:  pb.NewMetricsClient(conn),
		conn:    conn,
		agentIP: agentIP,
	}, nil
}

// Close закрывает gRPC-соединение.
func (s *GRPCMetricsSender) Close() error {
	return s.conn.Close()
}

// Send отправляет батч метрик на gRPC-сервер.
func (s *GRPCMetricsSender) Send(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	protoMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		b := pb.Metric_builder{Id: m.ID}
		switch m.MType {
		case models.Gauge:
			b.Type = pb.Metric_GAUGE
			if m.Value != nil {
				b.Value = *m.Value
			}
		case models.Counter:
			b.Type = pb.Metric_COUNTER
			if m.Delta != nil {
				b.Delta = *m.Delta
			}
		}
		protoMetrics = append(protoMetrics, b.Build())
	}

	if s.agentIP != "" {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("x-real-ip", s.agentIP))
	}

	req := pb.UpdateMetricsRequest_builder{Metrics: protoMetrics}.Build()
	_, err := s.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc UpdateMetrics: %w", err)
	}

	return nil
}

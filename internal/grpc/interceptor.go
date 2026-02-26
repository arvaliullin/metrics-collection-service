package grpc

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnetInterceptor возвращает UnaryServerInterceptor, проверяющий принадлежность IP агента доверенной подсети.
func TrustedSubnetInterceptor(trustedSubnet string, logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if trustedSubnet == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn().Msg("grpc: missing metadata")
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 || values[0] == "" {
			logger.Warn().Msg("grpc: missing x-real-ip metadata")
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		clientIP := net.ParseIP(values[0])
		if clientIP == nil {
			logger.Warn().Str("x_real_ip", values[0]).Msg("grpc: invalid x-real-ip")
			return nil, status.Error(codes.PermissionDenied, "invalid x-real-ip")
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Error().Err(err).Str("trusted_subnet", trustedSubnet).Msg("grpc: invalid trusted subnet CIDR")
			return nil, status.Error(codes.Internal, "invalid trusted subnet configuration")
		}

		if !ipNet.Contains(clientIP) {
			logger.Warn().Str("ip", values[0]).Str("trusted_subnet", trustedSubnet).Msg("grpc: IP not in trusted subnet")
			return nil, status.Error(codes.PermissionDenied, "IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}

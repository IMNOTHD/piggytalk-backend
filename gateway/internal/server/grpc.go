package server

import (
	eaV1 "gateway/api/event/v1"
	upV1 "gateway/api/upload/v1"
	"gateway/internal/conf"
	eventV1 "gateway/internal/service/event/v1"
	uploadV1 "gateway/internal/service/upload/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, eventV1 *eventV1.EventStreamService, uploadV1 *uploadV1.UploadService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	eaV1.RegisterEventStreamServer(srv, eventV1)
	upV1.RegisterUploadServer(srv, uploadV1)

	return srv
}

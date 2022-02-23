package server

import (
	accountV1 "account/api/account/v1"
	relationV1 "account/api/relation/v1"
	"account/internal/conf"
	accountV1Service "account/internal/service/account/v1"
	relationV1Service "account/internal/service/relation/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, account *accountV1Service.AccountService, friend *relationV1Service.FriendRelationService, logger log.Logger) *grpc.Server {
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
	accountV1.RegisterAccountServer(srv, account)
	relationV1.RegisterFriendRelationServer(srv, friend)

	return srv
}

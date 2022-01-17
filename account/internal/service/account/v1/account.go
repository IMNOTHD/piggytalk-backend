package service

import (
	"context"

	pb "account/api/account/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AccountService struct {
	pb.UnimplementedAccountServer
	log *log.Helper
}

func NewAccountService(logger log.Logger) *AccountService {
	return &AccountService{log: log.NewHelper(log.With(logger, "module", "account/service/account/v1", "caller", log.DefaultCaller))}
}

func (s *AccountService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{}, nil
}
func (s *AccountService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {

	return &pb.RegisterReply{}, nil
}

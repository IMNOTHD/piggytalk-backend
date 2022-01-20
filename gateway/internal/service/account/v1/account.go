package v1

import (
	"context"

	pb "gateway/api/account/v1"
	"gateway/internal/biz/account/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AccountService struct {
	pb.UnimplementedAccountServer

	au  *v1.AccountUsecase
	log *log.Helper
}

func NewAccountService(au *v1.AccountUsecase, logger log.Logger) *AccountService {
	return &AccountService{
		au:  au,
		log: log.NewHelper(log.With(logger, "module", "account/service/account/v1", "caller", log.DefaultCaller))}
}

func (s *AccountService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{}, nil
}
func (s *AccountService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	t, err := s.au.Register(ctx, &v1.Account{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Nickname: req.GetNickname(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Avatar:   req.GetAvatar(),
	})
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	return &pb.RegisterReply{Token: string(t)}, nil
}

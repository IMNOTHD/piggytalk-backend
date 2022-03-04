package service

import (
	"context"
	"fmt"

	pb "account/api/account/v1"
	"account/internal/biz/account/v1"

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
		log: log.NewHelper(log.With(logger, "module", "account/service/account/v1", "caller", log.DefaultCaller)),
	}
}

func (s *AccountService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	x, err := s.au.GetUserInfo(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

	var r []*pb.GetUserInfoResponse_UserInfo
	for _, info := range x {
		r = append(r, &pb.GetUserInfoResponse_UserInfo{
			Uuid:     info.Uuid,
			Avatar:   info.Avatar,
			Nickname: info.Nickname,
		})
	}

	return &pb.GetUserInfoResponse{Userinfo: r}, nil
}

func (s *AccountService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	a, t, err := s.au.Login(ctx, &v1.Account{
		Username: req.GetAccount(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	p := ""
	if a.Phone != "" {
		p = fmt.Sprintf("%s******%s", a.Phone[0:3], a.Phone[len(a.Phone)-2:])
	}

	return &pb.LoginReply{
		Token:    t.Token,
		Username: a.Username,
		Email:    a.Email,
		Phone:    p,
		Avatar:   a.Avatar,
		Nickname: a.Nickname,
	}, nil
}
func (s *AccountService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	t, err := s.au.CreateUser(ctx, &v1.Account{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Nickname: req.GetNickname(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Avatar:   req.GetAvatar(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterReply{Token: t.Token}, nil
}

func (s *AccountService) CheckLoginStat(ctx context.Context, req *pb.CheckLoginStatRequest) (*pb.CheckLoginStatResponse, error) {
	t, err := s.au.CheckUserLoginStatus(ctx, &v1.TokenInfo{Token: req.GetToken(), Device: v1.DeviceWeb})
	if err != nil {
		return nil, err
	}

	if t.Token == req.Token {
		switch t.Device {
		case v1.DeviceWeb:
			return &pb.CheckLoginStatResponse{
				Token:  t.Token,
				Device: pb.CheckLoginStatResponse_WEB,
				Uuid:   t.UserUUID.String(),
			}, nil
		case v1.DevicePhone:
			return &pb.CheckLoginStatResponse{
				Token:  t.Token,
				Device: pb.CheckLoginStatResponse_PHONE,
				Uuid:   t.UserUUID.String(),
			}, nil
		}
	}

	return &pb.CheckLoginStatResponse{Token: ""}, nil
}

package service

import (
	v1 "gateway/api/account/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AccountService struct {
	v1.UnimplementedAccountServer

	log *log.Helper
}

func NewAccountService(logger log.Logger) *AccountService {
	return &AccountService{log: log.NewHelper(logger)}
}

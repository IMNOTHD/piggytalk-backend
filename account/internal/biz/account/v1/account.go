package v1

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type Account struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Avatar   string
}

const (
	DeviceWeb   = "web"
	DevicePhone = "phone"
)

type TokenInfo struct {
	Token    string
	Device   string
	UserUUID *uuid.UUID
}

type AccountRepo interface {
	CreateUser(ctx context.Context, a *Account) (*uuid.UUID, error)
	CreateUserLoginToken(ctx context.Context, t *TokenInfo) (*TokenInfo, error)
}

type AccountUsecase struct {
	repo AccountRepo
	log  *log.Helper
}

func NewAccountUsecase(repo AccountRepo, logger log.Logger) *AccountUsecase {
	return &AccountUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "account/biz/account/v1", "caller", log.DefaultCaller)),
	}
}

func (uc *AccountUsecase) CreateUser(ctx context.Context, a *Account) (*TokenInfo, error) {
	x, err := uc.repo.CreateUser(ctx, a)
	if err != nil {
		return nil, err
	}

	return uc.repo.CreateUserLoginToken(ctx, &TokenInfo{
		Token:    "",
		Device:   DeviceWeb,
		UserUUID: x,
	})
}

func (uc *AccountUsecase) CheckUserLoginStatus(ctx context.Context, t *TokenInfo) (*TokenInfo, error) {
	return nil, nil
}

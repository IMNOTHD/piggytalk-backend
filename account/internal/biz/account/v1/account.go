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
	UUID     uuid.UUID
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
	CheckUserPassword(ctx context.Context, a *Account) (*Account, error)
	CheckToken(ctx context.Context, t *TokenInfo) (*TokenInfo, error)
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

func (uc *AccountUsecase) Login(ctx context.Context, a *Account) (*Account, *TokenInfo, error) {
	ac, err := uc.repo.CheckUserPassword(ctx, a)
	if err != nil {
		return nil, nil, err
	}

	t, err := uc.repo.CreateUserLoginToken(ctx, &TokenInfo{
		Token:    "",
		Device:   DeviceWeb,
		UserUUID: &ac.UUID,
	})

	return ac, t, err
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
	t, err := uc.repo.CheckToken(ctx, &TokenInfo{
		Token:  t.Token,
		Device: t.Device,
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

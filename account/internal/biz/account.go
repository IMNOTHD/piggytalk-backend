package biz

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

type TokenInfo struct {
	Token    string
	UserUUID *uuid.UUID
}

type AccountRepo interface {
	CreateUser(ctx context.Context, a *Account) (*uuid.UUID, error)
	CreateUserLoginToken(ctx context.Context, t *TokenInfo) (*TokenInfo, error)
}

type AccountUserCase struct {
	repo AccountRepo
	log  *log.Helper
}

func NewAccountUserCase(repo AccountRepo, logger log.Logger) *AccountUserCase {
	return &AccountUserCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "account/biz/account", "caller", log.DefaultCaller)),
	}
}

func (uc *AccountUserCase) CreateUser(ctx context.Context, a *Account) (*TokenInfo, error) {
	x, err := uc.repo.CreateUser(ctx, a)
	if err != nil {
		return nil, err
	}

	return uc.repo.CreateUserLoginToken(ctx, &TokenInfo{
		Token:    "",
		UserUUID: x,
	})
}

func (uc *AccountUserCase) CheckUserLoginStatus(ctx context.Context, t *TokenInfo) (*TokenInfo, error) {

}

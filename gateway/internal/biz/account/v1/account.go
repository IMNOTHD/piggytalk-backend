package v1

import (
	"context"
	"regexp"

	v1 "gateway/internal/api/account/account/v1"
	"gateway/internal/kit"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	_usernameExp = `^[a-zA-Z0-9_]{3,16}$`
	_emailExp    = `^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
	_phoneExp    = `^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$`
)

type AccountRepo interface {
}

type Token string

type Account struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Avatar   string
}

type AccountUsecase struct {
	repo AccountRepo
	log  *log.Helper
}

func NewAccountUsecase(repo AccountRepo, logger log.Logger) *AccountUsecase {
	return &AccountUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "gateway/biz/account/v1", "caller", log.DefaultCaller)),
	}
}

func (au *AccountUsecase) Register(ctx context.Context, a *Account) (Token, error) {
	if r := regexp.MustCompile(_usernameExp); a.Username == "" || !r.Match([]byte(a.Username)) {
		return "", errors.New(400, "BAD_REQUEST", "用户名格式错误")
	}
	if a.Password == "" {
		return "", errors.New(400, "BAD_REQUEST", "密码格式错误")
	}
	if r := regexp.MustCompile(_emailExp); a.Email != "" && !r.Match([]byte(a.Email)) {
		return "", errors.New(400, "BAD_REQUEST", "邮箱格式错误")
	}
	if r := regexp.MustCompile(_phoneExp); a.Phone != "" && !r.Match([]byte(a.Phone)) {
		return "", errors.New(400, "BAD_REQUEST", "手机格式错误")
	}

	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		au.log.Error(err)
		return "", err
	}
	c := v1.NewAccountClient(conn)

	ar, err := c.Register(ctx, &v1.RegisterRequest{
		Username: a.Username,
		Password: a.Password,
		Email:    a.Email,
		Phone:    a.Phone,
		Avatar:   a.Avatar,
		Nickname: a.Nickname,
	})
	if err != nil {
		au.log.Error(err)
		return "", err
	}

	return Token(ar.GetToken()), nil
}

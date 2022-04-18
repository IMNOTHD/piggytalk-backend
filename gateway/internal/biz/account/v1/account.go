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
	_usernameExp = `^[a-zA-Z][a-zA-Z0-9_]{2,15}$`
	_emailExp    = `^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
	_phoneExp    = `^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$`
)

type AccountRepo interface {
}

type Token string
type Uuid string

type Account struct {
	Username string
	Password string
	Nickname string
	Email    string
	Phone    string
	Avatar   string
	Uuid     string
}

type AccountUsecase struct {
	repo AccountRepo
	log  *log.Helper
}

func NewAccountUsecase(logger log.Logger) *AccountUsecase {
	return &AccountUsecase{
		log: log.NewHelper(log.With(logger, "module", "gateway/biz/account/v1", "caller", log.DefaultCaller)),
	}
}

func (au *AccountUsecase) SearchUuid(ctx context.Context, uuid string) (*Account, error) {
	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		au.log.Error(err)
		return nil, err
	}
	c := v1.NewAccountClient(conn)

	r, err := c.GetUserInfo(ctx, &v1.GetUserInfoRequest{Uuid: []string{uuid}})
	if err != nil {
		au.log.Error(err)
		return nil, err
	}

	if len(r.GetUserinfo()) != 0 {
		return &Account{
			Nickname: r.Userinfo[0].Nickname,
			Avatar:   r.Userinfo[0].Avatar,
		}, nil
	} else {
		return nil, nil
	}
}

func (au *AccountUsecase) UpdateAvatar(ctx context.Context, token string, avatar string) error {
	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		au.log.Error(err)
		return err
	}
	c := v1.NewAccountClient(conn)

	ar, err := c.UpdateAvatar(ctx, &v1.UpdateAvatarRequest{
		Token:  token,
		Avatar: avatar,
	})
	if err != nil {
		au.log.Error(err)
		return err
	}

	if ar.GetToken() == "" {
		au.log.Error("token failed")
		return errors.New(403, "FORBIDDEN", "FORBIDDEN")
	}

	return nil
}

func (au *AccountUsecase) Login(ctx context.Context, a *Account) (*Account, Token, error) {
	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		au.log.Error(err)
		return nil, "", err
	}
	c := v1.NewAccountClient(conn)

	ar, err := c.Login(ctx, &v1.LoginRequest{
		Account:  a.Username,
		Password: a.Password,
	})
	if err != nil {
		au.log.Error(err)
		return nil, "", err
	}

	return &Account{
		Username: ar.GetUsername(),
		Nickname: ar.GetNickname(),
		Email:    ar.Email,
		Phone:    ar.Phone,
		Avatar:   ar.GetAvatar(),
		Uuid:     ar.GetUuid(),
	}, Token(ar.GetToken()), nil
}

func (au *AccountUsecase) Register(ctx context.Context, a *Account) (Token, Uuid, error) {
	if r := regexp.MustCompile(_usernameExp); a.Username == "" || !r.Match([]byte(a.Username)) {
		return "", "", errors.New(400, "BAD_REQUEST", "用户名格式错误")
	}
	if a.Password == "" {
		return "", "", errors.New(400, "BAD_REQUEST", "密码格式错误")
	}
	if r := regexp.MustCompile(_emailExp); a.Email != "" && !r.Match([]byte(a.Email)) {
		return "", "", errors.New(400, "BAD_REQUEST", "邮箱格式错误")
	}
	if r := regexp.MustCompile(_phoneExp); a.Phone != "" && !r.Match([]byte(a.Phone)) {
		return "", "", errors.New(400, "BAD_REQUEST", "手机格式错误")
	}
	if len(a.Nickname) <= 0 || len(a.Nickname) > 32 {
		return "", "", errors.New(400, "BAD_REQUEST", "昵称过长")
	}

	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		au.log.Error(err)
		return "", "", err
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
		return "", "", err
	}

	return Token(ar.GetToken()), Uuid(ar.Uuid), nil
}

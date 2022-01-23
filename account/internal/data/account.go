package data

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"account/ent/user"
	"account/ent/userinfo"
	sv1 "account/internal/api/snowflake/snowflake/v1"
	"account/internal/biz/account/v1"
	"account/internal/kit"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type accountRepo struct {
	data *Data
	log  *log.Helper
}

// NewAccountRepo .
func NewAccountRepo(data *Data, logger log.Logger) v1.AccountRepo {
	return &accountRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "account/data/account/v1", "caller", log.DefaultCaller)),
	}
}

func (r *accountRepo) CheckUserPassword(ctx context.Context, a *v1.Account) (*v1.Account, error) {
	ux, err := r.data.Db.User.
		Query().
		Where(user.UsernameEQ(a.Username)).
		All(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	xx, err := r.data.Db.UserInfo.
		Query().
		Where(userinfo.Or(
			userinfo.PhoneEQ(a.Username),
			userinfo.EmailEQ(a.Email))).
		All(ctx)
	if len(ux) == 0 && len(xx) == 0 {
		r.log.Infof("user not exists: %s", a.Username)
		return nil, errors.New(400, "BAD_REQUEST", "用户名或密码错误")
	}

	var uid uuid.UUID
	if len(ux) != 0 {
		u := ux[0]
		uid = u.UUID
		//验证密码
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(a.Password+uid.String()))
		if err != nil {
			r.log.Infof("user password error: %s", a.Username)
			return nil, errors.New(400, "BAD_REQUEST", "用户名或密码错误")
		}

		x, err := r.data.Db.UserInfo.
			Query().
			Where(userinfo.UUIDEQ(uid)).
			First(ctx)
		if err != nil {
			r.log.Error(err)
			return nil, err
		}
		return &v1.Account{
			Username: u.Username,
			Nickname: x.Nickname,
			Email:    x.Email,
			Phone:    x.Phone,
			Avatar:   x.Avatar,
			UUID:     uid,
		}, nil
	} else {
		x := xx[0]
		uid = x.UUID
		u, err := r.data.Db.User.
			Query().
			Where(user.UUIDEQ(uid)).
			First(ctx)
		if err != nil {
			r.log.Error(err)
			return nil, err
		}
		// 验证密码
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(a.Password+uid.String()))
		if err != nil {
			r.log.Infof("user password error: %s", a.Username)
			return nil, errors.New(400, "BAD_REQUEST", "用户名或密码错误")
		}

		return &v1.Account{
			Username: u.Username,
			Nickname: x.Nickname,
			Email:    x.Email,
			Phone:    x.Phone,
			Avatar:   x.Avatar,
			UUID:     uid,
		}, nil
	}

}

func (r *accountRepo) CreateUserLoginToken(ctx context.Context, t *v1.TokenInfo) (*v1.TokenInfo, error) {
	conn, err := kit.ServiceConn(kit.SnowflakeEndpoint)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	c := sv1.NewSnowflakeClient(conn)
	sr, err := c.CreateSnowflake(ctx, &sv1.CreateSnowflakeRequest{
		DataCenterId: 0,
		WorkerId:     0,
	})
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	t.Token = strconv.Itoa(int(sr.GetSnowFlakeId()))

	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:account:uuid2token:")
	buffer.WriteString(t.UserUUID.String())
	buffer.WriteString(":")
	buffer.WriteString(t.Device)

	x := r.data.Rdb.SAdd(ctx, buffer.String(), t.Token)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return nil, x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:account:token2uuid:")
	buffer.WriteString(string(t.Device))
	buffer.WriteString(":")
	buffer.WriteString(t.Token)
	x = r.data.Rdb.SAdd(ctx, buffer.String(), t.UserUUID.String())
	if x.Err() != nil {
		r.log.Error(x.Err())
		return nil, x.Err()
	}

	return t, nil
}

func (r *accountRepo) CreateUser(ctx context.Context, a *v1.Account) (*uuid.UUID, error) {
	x, err := r.data.Db.User.
		Query().
		Where(user.UsernameEQ(a.Username)).
		All(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	if len(x) != 0 {
		r.log.Infof("username exists: %s", a.Username)
		return nil, errors.New(400, "BAD_REQUEST", "用户名已存在")
	}
	y, err := r.data.Db.UserInfo.
		Query().
		Where(userinfo.PhoneEQ(a.Phone)).
		All(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	if len(y) != 0 {
		r.log.Infof("phone exists: %s", a.Phone)
		return nil, errors.New(400, "BAD_REQUEST", "手机已被使用")
	}
	z, err := r.data.Db.UserInfo.
		Query().
		Where(userinfo.EmailEQ(a.Email)).
		All(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	if len(z) != 0 {
		r.log.Infof("email exists: %s", a.Email)
		return nil, errors.New(400, "BAD_REQUEST", "邮箱已被使用")
	}

	userUuid := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password+userUuid.String()), 10)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	u, err := r.data.Db.User.Create().
		SetUsername(a.Username).
		SetPassword(string(hash)).
		SetUUID(userUuid).
		SetGmtCreate(time.Now()).
		SetGmtModified(time.Now()).
		Save(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	r.log.Infof("success create user %s, id %d", u.Username, u.ID)

	ui, err := r.data.Db.UserInfo.Create().
		SetUUID(userUuid).
		SetGmtCreate(time.Now()).
		SetGmtModified(time.Now()).
		SetNickname(a.Nickname).
		SetEmail(a.Email).
		SetPhone(a.Phone).
		Save(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	r.log.Infof("success create userInfo, id %d", ui.ID)

	return &userUuid, nil
}

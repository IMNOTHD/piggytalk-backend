package v1

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"time"

	"account/ent/user"
	sv1 "account/internal/api/snowflake/snowflake/v1"
	"account/internal/biz/account/v1"
	d "account/internal/data"
	"account/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var _ v1.AccountRepo = (*accountRepo)(nil)

type accountRepo struct {
	data *d.Data
	log  *log.Helper
}

// NewAccountRepo .
func NewAccountRepo(data *d.Data, logger log.Logger) v1.AccountRepo {
	return &accountRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "account/data/account/v1", "caller", log.DefaultCaller)),
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
	userUuid := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), 0)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	x, err := r.data.Db.User.
		Query().
		Where(user.UsernameEQ(a.Username)).
		All(ctx)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}
	if x != nil {
		r.log.Infof("username exists: %s", a.Username)
		return nil, errors.New("username exists")
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

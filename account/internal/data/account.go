package data

import (
	"context"
	"errors"
	"time"

	"account/ent/user"
	"account/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var _ biz.AccountRepo = (*accountRepo)(nil)

type accountRepo struct {
	data *Data
	log  *log.Helper
}

// NewAccountRepo .
func NewAccountRepo(data *Data, logger log.Logger) biz.AccountRepo {
	return &accountRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "account/data/account", "caller", log.DefaultCaller)),
	}
}

func (r *accountRepo) CreateUserLoginToken(ctx context.Context, t *biz.TokenInfo) (*biz.TokenInfo, error) {

	err := r.data.rdb.Set(ctx)
}

func (r *accountRepo) CreateUser(ctx context.Context, a *biz.Account) (*uuid.UUID, error) {
	userUuid := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), 0)
	if err != nil {
		return nil, err
	}

	x, err := r.data.db.User.
		Query().
		Where(user.UsernameEQ(a.Username)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if x != nil {
		return nil, errors.New("username exists")
	}

	u, err := r.data.db.User.Create().
		SetUsername(a.Username).
		SetPassword(string(hash)).
		SetUUID(userUuid).
		SetGmtCreate(time.Now()).
		SetGmtModified(time.Now()).
		Save(ctx)
	r.log.Infof("success create user, id %d", u.ID)
	if err != nil {
		return nil, err
	}

	ui, err := r.data.db.UserInfo.Create().
		SetUUID(userUuid).
		SetGmtCreate(time.Now()).
		SetGmtModified(time.Now()).
		SetNickname(a.Nickname).
		SetEmail(a.Email).
		SetPhone(a.Phone).
		Save(ctx)
	r.log.Infof("success create userInfo, id %d", ui.ID)
	if err != nil {
		return nil, err
	}

	return &userUuid, nil
}

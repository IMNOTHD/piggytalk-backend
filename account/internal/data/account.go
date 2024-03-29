package data

import (
	"bytes"
	"context"
	"strconv"
	"time"

	sv1 "account/internal/api/snowflake/snowflake/v1"
	"account/internal/biz/account/v1"
	"account/internal/kit"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type accountRepo struct {
	data *Data
	log  *log.Helper
}

// NewAccountRepo .
func NewAccountRepo(data *Data, logger log.Logger) v1.AccountRepo {
	return &accountRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "account/data/account", "caller", log.DefaultCaller)),
	}
}

type User struct {
	ID        uint
	Username  string    `gorm:"unique;not null;index:idx_username"`
	Password  string    `gorm:"not null"`
	UUID      uuid.UUID `gorm:"not null;index:idx_uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserInfo struct {
	ID        uint
	UUID      uuid.UUID `gorm:"not null;index:idx_uuid"`
	Nickname  string    `gorm:"not null"`
	Avatar    string
	Email     string `gorm:"index:idx_email"`
	Phone     string `gorm:"index:idx_phone"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	_singleMessagePrefix = "single_message_"
	_groupMessagePrefix  = "group_message_"
)

// SingleMessage 每个用户有自己的消息保存表, 写扩散, 这里做创建表用, 表名`single_message_*uuid*`
type SingleMessage struct {
	MessageId  int64     `gorm:"primaryKey"`
	SenderUuid uuid.UUID `gorm:"not null;index:idx_sender"`
	// 表示在与谁聊天
	Talk        uuid.UUID `gorm:"not null;index:idx_talk"`
	Message     []byte
	MessageUuid uuid.UUID `gorm:"not null;index:idx_message_uuid"`
	AlreadyRead bool
	CreatedAt   time.Time
}

func (r *accountRepo) UpdateAvatar(ctx context.Context, uuid string, avatar string) error {
	ru := r.data.Db.Model(&UserInfo{}).Where("uuid = ?", uuid).Update("avatar", avatar)
	if ru.Error != nil {
		r.log.Error(ru.Error)
		return ru.Error
	}

	return nil
}

func (r *accountRepo) SelectUserInfo(ctx context.Context, uuids []string) ([]*v1.NoSecretUserInfo, error) {
	var userInfos []*UserInfo

	ru := r.data.Db.Where("uuid IN ?", uuids).Find(&userInfos)
	if ru.Error != nil && !errors.Is(ru.Error, gorm.ErrRecordNotFound) {
		r.log.Error(ru.Error)
		return nil, ru.Error
	}

	var x []*v1.NoSecretUserInfo
	for _, info := range userInfos {
		x = append(x, &v1.NoSecretUserInfo{
			Uuid:     info.UUID.String(),
			Avatar:   info.Avatar,
			Nickname: info.Nickname,
		})
	}

	return x, nil
}

func (r *accountRepo) CheckToken(ctx context.Context, t *v1.TokenInfo) (*v1.TokenInfo, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:account:token2uuid:")
	buffer.WriteString(string(t.Device))
	buffer.WriteString(":")
	buffer.WriteString(t.Token)

	v, err := r.data.Rdb.Get(ctx, buffer.String()).Result()
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	uid, err := uuid.Parse(v)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	return &v1.TokenInfo{
		Token:    t.Token,
		Device:   t.Device,
		UserUUID: &uid,
	}, nil
}

func (r *accountRepo) CheckUserPassword(ctx context.Context, a *v1.Account) (*v1.Account, error) {
	var u User
	var ui UserInfo

	ru := r.data.Db.
		Where(`uuid IN (select users.uuid from users where users.username=? union all select user_infos.uuid from user_infos where user_infos.phone=? union all select user_infos.uuid from user_infos where user_infos.email=?)`, a.Username, a.Username, a.Username).
		Find(&u)

	if ru.Error != nil && !errors.Is(ru.Error, gorm.ErrRecordNotFound) {
		r.log.Error(ru.Error)
		return nil, ru.Error
	}
	if ru.RowsAffected != 0 {
		// 验证密码
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(a.Password+u.UUID.String()))
		if err != nil {
			r.log.Infof("user password error: %s", a.Username)
			return nil, errors.New(400, "BAD_REQUEST", "用户名或密码错误")
		}

		if rui := r.data.Db.Where(&UserInfo{UUID: u.UUID}).First(&ui); rui.Error != nil {
			r.log.Error(rui.Error)
			return nil, rui.Error
		}

		r.log.Infof("%s success check password", u.UUID)
		return &v1.Account{
			Username: u.Username,
			Nickname: ui.Nickname,
			Email:    ui.Email,
			Phone:    ui.Phone,
			Avatar:   ui.Avatar,
			UUID:     u.UUID,
		}, nil
	}

	// 用户未找到
	r.log.Infof("user not exists: %s", a.Username)
	return nil, errors.New(400, "BAD_REQUEST", "用户名或密码错误")
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

	x := r.data.Rdb.Set(ctx, buffer.String(), t.Token, 0)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return nil, x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:account:token2uuid:")
	buffer.WriteString(string(t.Device))
	buffer.WriteString(":")
	buffer.WriteString(t.Token)
	x = r.data.Rdb.Set(ctx, buffer.String(), t.UserUUID.String(), 0)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return nil, x.Err()
	}

	r.log.Infof("%s success login in device %s", t.UserUUID, t.Device)
	return t, nil
}

func (r *accountRepo) CreateUser(ctx context.Context, a *v1.Account) (*uuid.UUID, error) {
	var u User
	var ui UserInfo

	// username查询
	ru := r.data.Db.Where(&User{
		Username: a.Username,
	}).Find(&u)
	if ru.Error != nil && !errors.Is(ru.Error, gorm.ErrRecordNotFound) {
		r.log.Error(ru.Error)
		return nil, ru.Error
	}
	if ru.RowsAffected != 0 {
		r.log.Errorf("username exists: %s", a.Username)
		return nil, errors.New(400, "BAD_REQUEST", "用户名已存在")
	}

	// phone查询
	rui := r.data.Db.Where(&UserInfo{
		Phone: a.Phone,
	}).First(&ui)
	if rui.Error != nil && !errors.Is(rui.Error, gorm.ErrRecordNotFound) {
		r.log.Error(rui.Error)
		return nil, rui.Error
	}
	if rui.RowsAffected != 0 {
		r.log.Infof("phone exists: %s", a.Phone)
		return nil, errors.New(400, "BAD_REQUEST", "手机不可用")
	}

	// email查询
	rui = r.data.Db.Where(&UserInfo{
		Email: a.Email,
	}).First(&ui)
	if rui.Error != nil && !errors.Is(rui.Error, gorm.ErrRecordNotFound) {
		r.log.Error(rui.Error)
		return nil, rui.Error
	}
	if rui.RowsAffected != 0 {
		r.log.Infof("email exists: %s", a.Email)
		return nil, errors.New(400, "BAD_REQUEST", "邮箱不可用")
	}

	userUuid := uuid.New()
	hash, err := bcrypt.GenerateFromPassword([]byte(a.Password+userUuid.String()), 10)
	if err != nil {
		r.log.Error(err)
		return nil, err
	}

	tErr := r.data.Db.Transaction(func(tx *gorm.DB) error {
		ru = tx.Create(&User{
			Username: a.Username,
			Password: string(hash),
			UUID:     userUuid,
		})
		if ru.Error != nil {
			r.log.Error(ru.Error)
			return ru.Error
		}

		ru = tx.Create(&UserInfo{
			UUID:     userUuid,
			Nickname: a.Nickname,
			Avatar:   "",
			Email:    a.Email,
			Phone:    a.Phone,
		})
		if ru.Error != nil {
			r.log.Error(ru.Error)
			return ru.Error
		}

		err := tx.Table(_singleMessagePrefix + userUuid.String()).Migrator().CreateTable(&SingleMessage{})
		if err != nil {
			r.log.Error(err)
			return err
		}
		//TODO group message table

		return nil
	})
	if tErr == nil {
		r.log.Infof("success create user %s, id %d", u.Username, u.ID)
		r.log.Infof("success create userInfo, id %d", ui.ID)
		r.log.Infof("success create user single message table, id %d", ui.ID)
	}

	return &userUuid, tErr
}

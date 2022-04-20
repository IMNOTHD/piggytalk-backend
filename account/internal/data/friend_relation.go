package data

import (
	"context"
	"time"

	v1 "account/internal/biz/relation/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type friendRelationRepo struct {
	data *Data
	log  *log.Helper
}

func NewFriendRelationRepo(data *Data, logger log.Logger) v1.FriendRelationRepo {
	return &friendRelationRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "account/data/friend_relation", "caller", log.DefaultCaller)),
	}
}

type FriendRelation struct {
	ID        uint
	UserA     uuid.UUID `gorm:"not null"`
	UserB     uuid.UUID `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *friendRelationRepo) CreateFriend(ctx context.Context, a uuid.UUID, b uuid.UUID) error {
	var f FriendRelation
	jump := false

	result := r.data.Db.Where(&FriendRelation{
		UserA: a,
		UserB: b,
	}).Find(&f)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.log.Error(result.Error)
		return result.Error
	}
	if result.RowsAffected != 0 {
		r.log.Infof("%s and %s already friend", a.String(), b.String())
		jump = true
	}

	if !jump {
		result = r.data.Db.Create(&FriendRelation{
			UserA: a,
			UserB: b,
		})
		if result.Error != nil {
			r.log.Error(result.Error)
			return result.Error
		}
	} else {
		jump = false
	}

	var ff FriendRelation
	result = r.data.Db.Where(&FriendRelation{
		UserA: b,
		UserB: a,
	}).Find(&ff)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.log.Error(result.Error)
		return result.Error
	}
	if result.RowsAffected != 0 {
		r.log.Infof("%s and %s already friend", a.String(), b.String())
		jump = true
	}

	if !jump {
		result = r.data.Db.Create(&FriendRelation{
			UserA: b,
			UserB: a,
		})
		if result.Error != nil {
			r.log.Error(result.Error)
			return result.Error
		}
	}

	return nil
}

func (r *friendRelationRepo) DeleteFriend(ctx context.Context, a uuid.UUID, b uuid.UUID) error {
	var f FriendRelation

	result := r.data.Db.Where(&FriendRelation{
		UserA: a,
		UserB: b,
	}).Delete(&f)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.log.Error(result.Error)
	}
	if result.RowsAffected == 0 {
		r.log.Warnf("%s and %s not friend", a.String(), b.String())
	}

	result = r.data.Db.Where(&FriendRelation{
		UserA: b,
		UserB: a,
	}).Delete(&f)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.log.Error(result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.log.Warnf("%s and %s not friend", a.String(), b.String())
	}

	return nil
}

func (r *friendRelationRepo) ListFriend(ctx context.Context, a uuid.UUID) ([]string, error) {
	var f []*FriendRelation

	result := r.data.Db.Where(&FriendRelation{
		UserA: a,
	}).Find(&f)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		r.log.Error(result.Error)
		return nil, result.Error
	}

	u := make([]string, 0)
	for _, v := range f {
		u = append(u, v.UserB.String())
	}

	return u, nil
}

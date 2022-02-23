package data

import (
	"context"

	"account/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAccountRepo, NewFriendRelationRepo)

// Data .
type Data struct {
	Db  *gorm.DB
	Rdb *redis.Client
	log *log.Helper
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)
	db, err := gorm.Open(mysql.Open(conf.Database.GetSource()), &gorm.Config{})
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	// 检验表格是否存在
	if !db.Migrator().HasTable(&User{}) {
		err = db.Migrator().CreateTable(&User{})
	}
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}
	if !db.Migrator().HasTable(&UserInfo{}) {
		err = db.Migrator().CreateTable(&UserInfo{})
	}
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}
	if !db.Migrator().HasTable(&FriendRelation{}) {
		err = db.Migrator().CreateTable(&FriendRelation{})
	}
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.GetAddr(),
		DB:           int(conf.Redis.GetDb()),
		ReadTimeout:  conf.Redis.GetReadTimeout().AsDuration(),
		WriteTimeout: conf.Redis.GetWriteTimeout().AsDuration(),
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		l.Errorf("redis ping check error: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
		l.Infof("closing the data resources")
		x, _ := db.DB()
		if err := x.Close(); err != nil {
			l.Error(err)
		}
		if err := rdb.Close(); err != nil {
			l.Error(err)
		}
	}
	return &Data{
		Db:  db,
		Rdb: rdb,
		log: l,
	}, cleanup, nil
}

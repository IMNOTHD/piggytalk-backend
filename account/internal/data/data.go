package data

import (
	"context"

	"account/ent"
	"account/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewAccountRepo)

// Data .
type Data struct {
	db  *ent.Client
	rdb *redis.Client
	log *log.Helper
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)
	db, err := ent.Open(conf.Database.GetDriver(), conf.Database.GetSource())
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}
	if err := db.Schema.Create(context.Background()); err != nil {
		l.Errorf("failed creating schema resources: %v", err)
		return nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.GetAddr(),
		DB:           int(conf.Redis.GetDb()),
		ReadTimeout:  conf.Redis.GetReadTimeout().AsDuration(),
		WriteTimeout: conf.Redis.GetWriteTimeout().AsDuration(),
	})

	cleanup := func() {
		l.Infof("closing the data resources")
		if err := db.Close(); err != nil {
			l.Error(err)
		}
		if err := rdb.Close(); err != nil {
			l.Error(err)
		}
	}
	return &Data{
		db:  db,
		rdb: rdb,
		log: l,
	}, cleanup, nil
}

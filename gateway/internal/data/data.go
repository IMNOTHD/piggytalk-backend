package data

import (
	"context"

	"gateway/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewEventRepo)

// Data .
type Data struct {
	Rdb *redis.Client
	log *log.Helper
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	rdb := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.GetAddr(),
		DB:           int(conf.Redis.GetDb()),
		ReadTimeout:  conf.Redis.GetReadTimeout().AsDuration(),
		WriteTimeout: conf.Redis.GetWriteTimeout().AsDuration(),
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		l.Errorf("redis ping check error: %v", err)
		return nil, nil, err
	}

	cleanup := func() {
		l.Infof("closing the data resources")
		if err := rdb.Close(); err != nil {
			l.Error(err)
		}
	}

	return &Data{
		Rdb: rdb,
		log: l,
	}, cleanup, nil
}

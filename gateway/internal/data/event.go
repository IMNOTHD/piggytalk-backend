package data

import (
	"bytes"
	"context"
	"time"

	v1 "gateway/internal/biz/event/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
)

type eventRepo struct {
	data *Data
	log  *log.Helper
}

func NewEventRepo(data *Data, logger log.Logger) v1.EventRepo {
	return &eventRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "gateway/data/event", "caller", log.DefaultCaller)),
	}
}

func (r *eventRepo) CreateSessionId(ctx context.Context, token string, sid string) (v1.SessionId, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2token:")
	buffer.WriteString(sid)

	x := r.data.Rdb.SAdd(ctx, buffer.String(), token)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:token2sessionId:")
	buffer.WriteString(token)
	x = r.data.Rdb.SAdd(ctx, buffer.String(), sid)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	return v1.SessionId(sid), nil
}

func (r *eventRepo) SelectToken(ctx context.Context, sessionId string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2token:")
	buffer.WriteString(sessionId)

	x, err := r.data.Rdb.Get(ctx, buffer.String()).Result()
	if err == redis.Nil {
		r.log.Errorf("sessionId %s not exists", sessionId)
		return "", err
	}
	if err != nil {
		r.log.Error(err)
		return "", err
	}

	return x, nil
}

func (r *eventRepo) RemoveSessionId(ctx context.Context, token string, sid string) error {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2token:")
	buffer.WriteString(sid)

	x := r.data.Rdb.Del(ctx, buffer.String())
	if x.Err() != nil {
		r.log.Error(x.Err())
		return x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:token2sessionId:")
	buffer.WriteString(token)
	x = r.data.Rdb.Del(ctx, buffer.String())
	if x.Err() != nil {
		r.log.Error(x.Err())
		return x.Err()
	}

	return nil
}

func (r *eventRepo) UpdateBeatHeart(ctx context.Context, sessionId string, expiration time.Duration) error {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:beatheart:")
	buffer.WriteString(sessionId)

	t := time.Now().UnixMilli()
	x := r.data.Rdb.SetEX(ctx, buffer.String(), t, expiration)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return x.Err()
	}

	return nil
}

func (r eventRepo) SelectBeatHeart(ctx context.Context, sessionId string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:beatheart:")
	buffer.WriteString(sessionId)

	x, err := r.data.Rdb.Get(ctx, buffer.String()).Result()
	if err == redis.Nil {
		r.log.Infof("session %s not exists", sessionId)
		return "", nil
	}
	if err != nil {
		r.log.Error(err)
		return "", err
	}

	return x, nil
}

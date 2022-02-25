package data

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	v1 "gateway/internal/biz/event/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
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

func (r *eventRepo) SendAddFriend(ctx context.Context, sid int64, receiverUuid string, note string, uid string) error {
	type body struct {
		ReceiverUuid string
		Note         string
		Uid          string
	}
	b := body{
		ReceiverUuid: receiverUuid,
		Note:         note,
		Uid:          uid,
	}
	x, err := json.Marshal(b)
	if err != nil {
		r.log.Error(err)
		return err
	}

	err = r.data.Rmq.Channel.Publish(
		_eventTopicEx,
		_eventMasterMQ,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			MessageId:   strconv.Itoa(int(sid)),
			Body:        x,
			Type:        "AddFriend",
		},
	)
	if err != nil {
		r.log.Error(err)
	}

	return err
}

func (r *eventRepo) CreateSessionId(ctx context.Context, token string, sid string, uid string) (v1.SessionId, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2token:")
	buffer.WriteString(sid)

	x := r.data.Rdb.Set(ctx, buffer.String(), token, 0)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:token2sessionId:")
	buffer.WriteString(token)
	x = r.data.Rdb.Set(ctx, buffer.String(), sid, 0)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:uid2sessionId:")
	buffer.WriteString(uid)
	x = r.data.Rdb.Set(ctx, buffer.String(), sid, 0)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:sessionId2uid:")
	buffer.WriteString(sid)
	x = r.data.Rdb.Set(ctx, buffer.String(), uid, 0)
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

func (r *eventRepo) SelectUid(ctx context.Context, sessionId string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2uid:")
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

	uid, err := r.SelectUid(ctx, sid)
	if err != nil {
		r.log.Error(err)
		return err
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:uid2sessionId:")
	buffer.WriteString(uid)
	x = r.data.Rdb.Del(ctx, buffer.String())
	if x.Err() != nil {
		r.log.Error(x.Err())
		return x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:gateway:sessionId2uid:")
	buffer.WriteString(sid)
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

package data

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	v1 "gateway/internal/biz/event/v1"
	con "gateway/internal/conf"
	v12 "gateway/internal/service/event/v1"

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

// message type
const ()

// event type
const (
	_addFriend    = "AddFriend"
	_createFriend = "CreateFriend"
	_deleteFriend = "DeleteFriend"
)

func (r *eventRepo) RabbitMqLister(ctx context.Context) (func(), func()) {
	eventSessionQueue := strconv.Itoa(int(con.WorkerId)) + _eventMQSuffix
	messageSessionQueue := strconv.Itoa(int(con.WorkerId)) + _messageMQSuffix

	// message session消息
	messageListener := func() {
		msg, err := r.data.Rmq.Channel.Consume(
			messageSessionQueue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			r.log.Errorf("Fail to register message consumer: %v", err)
		}

		for m := range msg {
			r.log.Infof("Receive Message: type: %s\nbody: %s\n correlation-id: %s\n message-id", m.Type, m.Body, m.CorrelationId, m.MessageId)
			if m.CorrelationId != "" && v12.ReceiveMessageMq[m.CorrelationId] != nil {
				v12.ReceiveMessageMq[m.CorrelationId] <- v12.Message{
					Type:      m.Type,
					Body:      string(m.Body),
					MessageId: m.MessageId,
				}
			}
		}
	}
	// 消费event session消息
	eventListener := func() {
		msg, err := r.data.Rmq.Channel.Consume(
			eventSessionQueue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			r.log.Errorf("Fail to register event consumer: %v", err)
		}

		for m := range msg {
			r.log.Infof("Receive Event: type: %s\nbody: %s\n correlation-id: %s\n message-id", m.Type, m.Body, m.CorrelationId, m.MessageId)
			if m.CorrelationId != "" && v12.ReceiveEventMq[m.CorrelationId] != nil {
				v12.ReceiveEventMq[m.CorrelationId] <- v12.Message{
					Type:      m.Type,
					Body:      string(m.Body),
					MessageId: m.MessageId,
				}
			}
		}
	}

	return messageListener, eventListener
}

func (r *eventRepo) SendDeleteFriend(ctx context.Context, uid string, deleteUuid string, sid int64, eventUuid string) error {
	type body struct {
		DeleteUuid string
		Uid        string
		EventUuid  string
	}
	b := body{
		DeleteUuid: deleteUuid,
		Uid:        uid,
		EventUuid:  eventUuid,
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
			Type:        _deleteFriend,
		},
	)
	if err != nil {
		r.log.Error(err)
	}

	return err
}

func (r *eventRepo) SendConfirmFriend(ctx context.Context, sid int64, receiverUuid string, uid string, eventUuid string, addStat string) error {
	type body struct {
		ReceiverUuid string
		Uid          string
		AddStatCode  string
		EventUuid    string
	}
	b := body{
		ReceiverUuid: receiverUuid,
		Uid:          uid,
		AddStatCode:  addStat,
		EventUuid:    eventUuid,
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
			Type:        _createFriend,
		},
	)
	if err != nil {
		r.log.Error(err)
	}

	return err
}

func (r *eventRepo) SendAddFriend(ctx context.Context, sid int64, receiverUuid string, note string, uid string, eventUuid string) error {
	type body struct {
		ReceiverUuid string
		Note         string
		Uid          string
		EventUuid    string
	}
	b := body{
		ReceiverUuid: receiverUuid,
		Note:         note,
		Uid:          uid,
		EventUuid:    eventUuid,
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
			Type:        _addFriend,
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

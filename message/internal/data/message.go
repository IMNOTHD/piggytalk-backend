package data

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	v1 "message/internal/biz/message/v1"
	"message/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
)

type messageRepo struct {
	data *Data
	log  *log.Helper
}

var consulClient *api.Client

// message type
const ()

// event type
const (
	_addFriend    = "AddFriend"
	_createFriend = "CreateFriend"
)

func NewMessageRepo(data *Data, logger log.Logger) v1.MessageRepo {
	c, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	})
	if err != nil {
		panic(err)
	}
	consulClient = c

	return &messageRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "message/data/message", "caller", log.DefaultCaller)),
	}
}

type FriendAddMessage struct {
	ID        uint
	MessageId int64
	UserA     uuid.UUID `gorm:"not null"`
	UserB     uuid.UUID `gorm:"not null"`
	Type      string    `gorm:"type:enum('Request', 'Allow', 'Delete');default:'Request'"`
	Ack       string    `gorm:"type:enum('FALSE', 'TRUE');default:'FALSE'"`
	EventUuid string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *messageRepo) RabbitMqLister(ctx context.Context) (func(), func()) {
	// 往回发消息一定要把message id发回去！

	// message消息
	messageListener := func() {
		msg, err := r.data.Rmq.Channel.Consume(
			_messageMQ,
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
			r.log.Infof("type: %s\nbody: %s", m.Type, m.Body)
		}
	}
	// 消费event消息
	eventListener := func() {
		msg, err := r.data.Rmq.Channel.Consume(
			_eventMQ,
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
			r.log.Infof("type: %s\nbody: %s", m.Type, m.Body)
			switch string(m.Body) {
			case _addFriend:
				err := r.AddFriend(ctx, m.Body, m.MessageId)
				if err != nil {
					r.log.Error(err)
				}
			}
		}
	}

	return messageListener, eventListener
}

func (r *messageRepo) AddFriend(ctx context.Context, body []byte, mid string) error {
	type b struct {
		ReceiverUuid string
		Note         string
		Uid          string
		EventUuid    string
	}

	var x b
	err := json.Unmarshal(body, &x)
	if err != nil {
		r.log.Error(err)
		return err
	}

	m, err := strconv.ParseInt(mid, 10, 64)
	if err != nil {
		r.log.Error(err)
		return err
	}

	ru := r.data.Db.Create(&FriendAddMessage{
		MessageId: m,
		UserA:     uuid.MustParse(x.Uid),
		UserB:     uuid.MustParse(x.ReceiverUuid),
		Type:      "Request",
		EventUuid: x.EventUuid,
	})
	if ru.Error != nil {
		r.log.Error(err)
		return err
	}
	r.log.Infof("Success insert add friend request: %s", body)

	// 已经成功入库, 下面给session发消息的错误不需要强制返回错误, 仅记录即可
	sid, err := r.selectUuidFromSession(ctx, x.ReceiverUuid)
	if err != nil {
		r.log.Error(err)
		return nil
	}
	s, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		r.log.Error(err)
		return nil
	}

	_, w := kit.GetDeviceID(s)
	kv, _, err := consulClient.KV().Get(strconv.Itoa(int(w)), nil)
	if err != nil {
		r.log.Error(err)
		return nil
	}
	eventSessionQueue := string(kv.Value) + _eventMQSuffix
	err = r.data.Rmq.Channel.Publish(
		_eventTopicEx,
		eventSessionQueue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
			Type:        _addFriend,
			MessageId:   mid,
		},
	)
	if err != nil {
		r.log.Error(err)
		return nil
	}

	return nil
}

func (r *messageRepo) selectUuidFromSession(ctx context.Context, uid string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:gateway:sessionId2uid:")
	buffer.WriteString(uid)

	x, err := r.data.Rdb.Get(ctx, buffer.String()).Result()
	if err == redis.Nil {
		r.log.Infof("uid %s not online", uid)
		return "", err
	}
	if err != nil {
		r.log.Error(err)
		return "", err
	}

	return x, nil
}

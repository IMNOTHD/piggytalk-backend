package data

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"strconv"
	"time"

	v1 "message/internal/biz/message/v1"
	"message/internal/kit"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type messageRepo struct {
	data *Data
	log  *log.Helper
}

var consulClient *api.Client

// message type
const (
	_singleTalk = "SingleTalk"
	_groupTalk  = "GroupTalk"
)

// event type
const (
	_addFriend        = "AddFriend"
	_createFriend     = "CreateFriend"
	_deleteFriend     = "DeleteFriend"
	_ackFriendMessage = "AckFriendMessage"
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
	EventId int64 `gorm:"primaryKey"`
	// UserA Sender
	UserA uuid.UUID `gorm:"not null;index:idx_sender"`
	// UserB Receiver
	UserB     uuid.UUID `gorm:"not null;index:idx_receiver"`
	Type      string    `gorm:"type:enum('WAITING', 'SUCCESS', 'DENIED');default:'WAITING'"`
	Ack       string    `gorm:"type:enum('FALSE', 'TRUE');default:'FALSE'"`
	EventUuid uuid.UUID `gorm:"index:idx_event_uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

const (
	_friendAddEventPrefix = "friend_add_"
	_groupAddEventPrefix  = "group_add_"
	_singleMessagePrefix  = "single_message_"
	_groupMessagePrefix   = "group_message_"
)

// SingleMessage 每个用户有自己的消息保存表, 写扩散
type SingleMessage struct {
	MessageId  int64     `gorm:"primaryKey"`
	SenderUuid uuid.UUID `gorm:"not null;index:idx_sender"`
	// 表示在与谁聊天
	Talk        uuid.UUID `gorm:"not null;index:idx_talk"`
	Message     string
	MessageUuid uuid.UUID `gorm:"index:idx_message_uuid"`
	AlreadyRead bool
	CreatedAt   time.Time
}

func (r *messageRepo) RabbitMqLister(ctx context.Context) (func(), func()) {
	// 往回发消息一定要把message-id发回去！
	// 往回发消息一定要把接收者uuid写在correlation-id发回去！

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
			switch m.Type {
			case _singleTalk:
				err := r.SingleMessage(ctx, m.Body, m.MessageId)
				if err != nil {
					r.log.Error(err)
				}
			}
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
			switch m.Type {
			case _addFriend:
				err := r.AddFriend(ctx, m.Body, m.MessageId)
				if err != nil {
					r.log.Error(err)
				}
			case _ackFriendMessage:
				err := r.AckFriendRequest(ctx, m.Body)
				if err != nil {
					r.log.Error(err)
				}
			}
		}
	}

	return messageListener, eventListener
}

func (r *messageRepo) AckFriendRequest(ctx context.Context, body []byte) error {
	type b struct {
		Uid     string
		EventId int64
	}
	var x b
	err := json.Unmarshal(body, &x)
	if err != nil {
		r.log.Error(err)
		return err
	}

	ru := r.data.Db.Where("event_id = ? and user_b = ?", x.EventId, uuid.MustParse(x.Uid)).Update("ack", "TRUE")
	if ru.Error != nil {
		r.log.Error(ru.Error)
		return ru.Error
	}

	return nil
}

func (r *messageRepo) ListFriendRequest(ctx context.Context, uuid string, startId int64, count int64) ([]*v1.FriendAddMessage, error) {
	if count <= 0 {
		count = math.MaxInt64
	}

	var fm []*FriendAddMessage
	ru := r.data.Db.
		Raw("select `event_id`, `user_a`, `user_b`, `type`, `ack`, `event_uuid` from `friend_add_messages` where `event_id` < ? and (`user_a` = ? or `user_b` = ?) order by `event_id` desc limit ?",
			startId, uuid, uuid, count).Scan(&fm)
	if ru.Error != nil && !errors.Is(ru.Error, gorm.ErrRecordNotFound) {
		r.log.Error(ru.Error)
		return nil, ru.Error
	}

	var k []*v1.FriendAddMessage
	for _, message := range fm {
		k = append(k, &v1.FriendAddMessage{
			EventUuid:    message.EventUuid.String(),
			EventId:      message.EventId,
			Ack:          message.Ack,
			ReceiverUuid: message.UserB.String(),
			SenderUuid:   message.UserA.String(),
			Type:         message.Type,
		})
	}

	return k, nil
}

func (r *messageRepo) SingleMessage(ctx context.Context, body []byte, mid string) error {
	type b struct {
		Talk        string
		SenderUuid  string
		Message     string
		MessageUuid string
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

	// 事务, 写扩散
	ru := r.data.Db.Transaction(func(tx *gorm.DB) error {
		// 发送者消息记录
		senderTableName := _singleMessagePrefix + x.SenderUuid
		a := tx.Table(senderTableName).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_uuid"}},
			DoNothing: true,
		}).Create(&SingleMessage{
			MessageId:   m,
			SenderUuid:  uuid.MustParse(x.SenderUuid),
			Talk:        uuid.MustParse(x.Talk),
			Message:     x.Message,
			MessageUuid: uuid.MustParse(x.MessageUuid),
			AlreadyRead: true,
		})
		//tx.Exec("INSERT INTO ? (`message_id`, `sender_uuid`, `talk`, `message`, `message_uuid`, `already_read`) SELECT ?, ?, ?, ?, ?, ? FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM ? WHERE `message_uuid` = ?)",
		//	senderTableName, m, uuid.MustParse(x.SenderUuid), uuid.MustParse(x.Talk), x.Message, uuid.MustParse(x.MessageUuid), true, uuid.MustParse(x.MessageUuid))
		if a.Error != nil {
			r.log.Error(a.Error)
			return a.Error
		}

		// 接收者消息记录
		talkTableName := _singleMessagePrefix + x.Talk
		b := tx.Table(talkTableName).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_uuid"}},
			DoNothing: true,
		}).Create(&SingleMessage{
			MessageId:   m,
			SenderUuid:  uuid.MustParse(x.SenderUuid),
			Talk:        uuid.MustParse(x.SenderUuid),
			Message:     x.Message,
			MessageUuid: uuid.MustParse(x.MessageUuid),
			AlreadyRead: false,
		})
		//tx.Exec("INSERT INTO ? (`message_id`, `sender_uuid`, `talk`, `message`, `message_uuid`, `already_read`) SELECT ?, ?, ?, ?, ?, ? FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM ? WHERE `message_uuid` = ?)",
		//	talkTableName, m, uuid.MustParse(x.SenderUuid), uuid.MustParse(x.Talk), x.Message, uuid.MustParse(x.MessageUuid), true, uuid.MustParse(x.MessageUuid))
		if b.Error != nil {
			r.log.Error(b.Error)
			return b.Error
		}

		return nil
	})
	if ru != nil {
		return ru
	}

	// 消息写入成功, 投递
	sid, err := r.selectUuidFromSession(ctx, x.Talk)
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
	messageSessionQueue := string(kv.Value) + _messageMQSuffix
	err = r.data.Rmq.Channel.Publish(
		_messageTopicEx,
		messageSessionQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			Body:          body,
			Type:          _singleTalk,
			MessageId:     mid,
			CorrelationId: x.Talk,
		},
	)
	if err != nil {
		r.log.Error(err)
		return nil
	}

	return nil
}

func (r *messageRepo) SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error) {
	var f FriendAddMessage

	ru := r.data.Db.Where(&FriendAddMessage{EventUuid: uuid.MustParse(eventUuid)}).Last(&f)
	if ru.Error != nil {
		r.log.Error(ru.Error)
		return "", "", ru.Error
	}

	return f.UserA.String(), f.UserB.String(), nil
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

	ru := r.data.Db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_uuid"}},
		DoNothing: true,
	}).Create(&FriendAddMessage{
		EventId:   m,
		UserA:     uuid.MustParse(x.Uid),
		UserB:     uuid.MustParse(x.ReceiverUuid),
		Type:      "WAITING",
		EventUuid: uuid.MustParse(x.EventUuid),
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
			ContentType:   "text/plain",
			Body:          body,
			Type:          _addFriend,
			MessageId:     mid,
			CorrelationId: x.ReceiverUuid,
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

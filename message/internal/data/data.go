package data

import (
	"context"

	"message/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/streadway/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewMessageRepo)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Data .
type Data struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Rmq *RabbitMQ
	log *log.Helper
}

const (
	// _eventMQ 接收所有event消息, 再存档并分发给对应的gateway
	_eventMQ          = "event-mq"
	_eventTopicEx     = "event-topic"
	_eventDeadTopicEx = "dead.event-topic"
	_eventDeadMQ      = "dead.event-mq"
	_eventMasterMQ    = "master.event-mq"
	_eventMQSuffix    = ".event-mq"
	// _messageMQ 接收所有message消息, 再存档并分发给对应的gateway
	_messageMQ          = "message-mq"
	_messageTopicEx     = "message-topic"
	_messageDeadTopicEx = "dead.message-topic"
	_messageDeadMQ      = "dead.message-mq"
	_messageMasterMQ    = "master.message-mq"
	_messageMQSuffix    = ".message-mq"
)

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

	db, err := gorm.Open(mysql.Open(conf.Database.GetSource()), &gorm.Config{})
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	// 检验表格是否存在
	if !db.Migrator().HasTable(&FriendAddMessage{}) {
		err = db.Migrator().CreateTable(&FriendAddMessage{})
	}
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	conn, err := amqp.Dial("amqp://" + conf.Rabbitmq.GetUser() +
		":" + conf.Rabbitmq.GetPassword() +
		"@" + conf.Rabbitmq.GetAddr())
	if err != nil {
		l.Errorf("rabbitmq connection error: %v", err)
		return nil, nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		l.Errorf("rabbitmq new channel error: %v", err)
		return nil, nil, err
	}

	// message topic交换机
	err = channel.ExchangeDeclare(
		_messageTopicEx,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq declare message exchange error: %v", err)
		return nil, nil, err
	}

	// message发送队列与绑定
	_, err = channel.QueueDeclare(
		_messageMQ,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq declare message queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		_messageMQ,
		_messageMasterMQ,
		_messageTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind message queue error: %v", err)
		return nil, nil, err
	}

	// message session死信交换机
	err = channel.ExchangeDeclare(
		_messageDeadTopicEx,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq declare message dead exchange error: %v", err)
		return nil, nil, err
	}

	messageDeadProps := make(map[string]interface{})
	// 死信队列中的消息, 30分钟后仍然投递失败, 则彻底删除
	messageDeadProps["x-message-ttl"] = 1800000
	// message session死信队列与绑定
	_, err = channel.QueueDeclare(
		_messageDeadMQ,
		false,
		false,
		false,
		false,
		messageDeadProps,
	)
	if err != nil {
		l.Errorf("rabbitmq declare message dead queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		_messageDeadMQ,
		"*"+_messageMQSuffix,
		_messageDeadTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind message dead queue error: %v", err)
		return nil, nil, err
	}

	// event topic交换机
	err = channel.ExchangeDeclare(
		_eventTopicEx,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)

	// event发送队列与绑定
	_, err = channel.QueueDeclare(
		_eventMQ,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq declare event queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		_eventMQ,
		_eventMasterMQ,
		_eventTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind event queue error: %v", err)
		return nil, nil, err
	}
	// event session死信交换机
	err = channel.ExchangeDeclare(
		_eventDeadTopicEx,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq declare event dead exchange error: %v", err)
		return nil, nil, err
	}

	eventDeadProps := make(map[string]interface{})
	// 死信队列中的消息, 30分钟后仍然投递失败, 则彻底删除
	eventDeadProps["x-message-ttl"] = 1800000
	// event session死信队列与绑定
	_, err = channel.QueueDeclare(
		_eventDeadMQ,
		false,
		false,
		false,
		false,
		eventDeadProps,
	)
	if err != nil {
		l.Errorf("rabbitmq declare event dead queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		_eventDeadMQ,
		"*"+_eventMQSuffix,
		_eventDeadTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind event dead queue error: %v", err)
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
		if err := conn.Close(); err != nil {
			l.Error(err)
		}
	}
	return &Data{
		Db:  db,
		Rdb: rdb,
		Rmq: &RabbitMQ{
			Conn:    conn,
			Channel: channel,
		},
		log: l,
	}, cleanup, nil
}

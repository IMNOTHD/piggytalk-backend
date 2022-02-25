package data

import (
	"context"

	"gateway/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/streadway/amqp"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewEventRepo)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Data .
type Data struct {
	Rdb *redis.Client
	Rmq *RabbitMQ
	log *log.Helper
}

const (
	// _eventMQ 接收所有event消息, 再存档并分发给对应的gateway
	_eventMQ       = "mevent-mq"
	_eventTopicEx  = "event-topic"
	_eventMasterMQ = "master.event"
	// _messageMQ 接收所有message消息, 再存档并分发给对应的gateway
	_messageMQ       = "message-mq"
	_messageTopicEx  = "message-topic"
	_messageMasterMQ = "master.message"
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
	// message发送队列
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
	// bind message queue to exchange
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
	// event发送队列
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
	// bind event queue to exchange
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

	cleanup := func() {
		l.Infof("closing the data resources")
		if err := rdb.Close(); err != nil {
			l.Error(err)
		}
		if err := conn.Close(); err != nil {
			l.Error(err)
		}
	}

	return &Data{
		Rdb: rdb,
		Rmq: &RabbitMQ{
			Conn:    conn,
			Channel: channel,
		},
		log: l,
	}, cleanup, nil
}

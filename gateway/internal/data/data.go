package data

import (
	"context"
	"strconv"

	con "gateway/internal/conf"

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
func NewData(conf *con.Data, logger log.Logger) (*Data, func(), error) {
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

	eventSessionQueue := strconv.Itoa(int(con.WorkerId)) + _eventMQSuffix
	messageSessionQueue := strconv.Itoa(int(con.WorkerId)) + _messageMQSuffix

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

	messageSessionProps := make(map[string]interface{})
	// 5分钟过期, 死信队列
	messageSessionProps["x-message-ttl"] = 300000
	messageSessionProps["x-dead-letter-exchange"] = _messageDeadTopicEx
	// message 本机session接收队列与绑定
	_, err = channel.QueueDeclare(
		eventSessionQueue,
		false,
		true,
		false,
		false,
		messageSessionProps,
	)
	if err != nil {
		l.Errorf("rabbitmq declare session message queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		messageSessionQueue,
		messageSessionQueue,
		_messageTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind session message queue error: %v", err)
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

	eventSessionProps := make(map[string]interface{})
	// 5分钟过期, 死信队列
	eventSessionProps["x-message-ttl"] = 300000
	eventSessionProps["x-dead-letter-exchange"] = _eventDeadTopicEx
	// event 本机session接收队列与绑定
	_, err = channel.QueueDeclare(
		strconv.Itoa(int(con.WorkerId))+_eventMQSuffix,
		false,
		true,
		false,
		false,
		eventSessionProps,
	)
	if err != nil {
		l.Errorf("rabbitmq declare session event queue error: %v", err)
		return nil, nil, err
	}
	err = channel.QueueBind(
		eventSessionQueue,
		eventSessionQueue,
		_eventTopicEx,
		false,
		nil,
	)
	if err != nil {
		l.Errorf("rabbitmq bind session event queue error: %v", err)
		return nil, nil, err
	}

	// 消费event session消息
	go func() {
		msg, err := channel.Consume(
			eventSessionQueue,
			con.ID.String(),
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			l.Errorf("Fail to register consumer: %v", err)
		}

		for m := range msg {
			l.Infof("type: %s\nbody: %s", m.Type, m.Body)
		}
	}()

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

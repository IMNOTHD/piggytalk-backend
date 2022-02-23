package data

import (
	"message/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/streadway/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// Data .
type Data struct {
	Db  *gorm.DB
	Rmq *RabbitMQ
	log *log.Helper
}

// NewData .
func NewData(conf *conf.Data, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(logger)

	db, err := gorm.Open(mysql.Open(conf.Database.GetSource()), &gorm.Config{})
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	// 检验表格是否存在
	// TODO

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

	cleanup := func() {
		l.Infof("closing the data resources")
		x, _ := db.DB()
		if err := x.Close(); err != nil {
			l.Error(err)
		}
		if err := conn.Close(); err != nil {
			l.Error(err)
		}
	}
	return &Data{
		Db: db,
		Rmq: &RabbitMQ{
			Conn:    conn,
			Channel: channel,
		},
		log: l,
	}, cleanup, nil
}

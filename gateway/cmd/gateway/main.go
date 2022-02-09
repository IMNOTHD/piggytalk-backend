package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"

	"gateway/internal/conf"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/contrib/log/fluent/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/consul/api"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	ConsulClient *api.Client

	// ID和WorkerId已被移入conf\global.go
)

const (
	// Name is the name of the compiled software.
	Name = "piggytalk-backend-gateway"
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")

	c, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	})
	if err != nil {
		panic(err)
	}
	ConsulClient = c
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(conf.ID.String()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
		kratos.Registrar(consul.New(ConsulClient)),
	)
}

func main() {
	fluentdService := ServiceDiscover("fluentd1")

	logger, err := fluent.NewLogger(
		fmt.Sprintf("tcp://%s:%d", "127.0.0.1", fluentdService.Port),
		fluent.WithTagPrefix("piggytalk-backend-gateway"))
	if err != nil {
		panic(err)
	}

	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 在consul的kv中注册workerId
	lock, err := ConsulClient.LockOpts(&api.LockOptions{
		Key:         "workerIdGeneratorLocker",
		Value:       []byte(conf.ID.String()),
		SessionName: conf.ID.String(),
		SessionTTL:  "10s",
	})
	if err != nil {
		panic(err)
	}
	ch := make(chan struct{})
	_, err = lock.Lock(ch)
	if err != nil {
		panic(err)
	}
	conf.WorkerId = uint(rand.Intn(1 << 7))
	kv, _, err := ConsulClient.KV().Get(strconv.Itoa(int(conf.WorkerId)), nil)
	if err != nil {
		_ = lock.Unlock()
		panic(err)
	}
	for kv != nil {
		conf.WorkerId = uint(rand.Intn(1 << 7))
		kv, _, err = ConsulClient.KV().Get(strconv.Itoa(int(conf.WorkerId)), nil)
		if err != nil {
			fmt.Println(err)
			_ = lock.Unlock()
			return
		}
	}
	_, err = ConsulClient.KV().Put(&api.KVPair{
		Key:   strconv.Itoa(int(conf.WorkerId)),
		Value: []byte(conf.ID.String()),
	}, nil)
	_, err = ConsulClient.KV().Put(&api.KVPair{
		Key:   conf.ID.String(),
		Value: []byte(strconv.Itoa(int(conf.WorkerId))),
	}, nil)
	if err != nil {
		_ = lock.Unlock()
		panic(err)
	}
	_ = logger.Log(log.LevelInfo, Name, fmt.Sprintf("%s %s success register workerId %d", Name, conf.ID.String(), conf.WorkerId))

	close(ch)
	err = lock.Unlock()
	if err != nil {
		fmt.Println(err)
	}

	app, cleanup, err := initApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	defer func() {
		_, err := ConsulClient.KV().Delete(strconv.Itoa(int(conf.WorkerId)), nil)
		if err != nil {
			fmt.Println(err)
		}
		_, err = ConsulClient.KV().Delete(conf.ID.String(), nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	_ = logger.Log(log.LevelInfo, Name, fmt.Sprintf("%s is ready to start...", conf.ID.String()))

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// ServiceDiscover 服务发现，获取指定id的服务
func ServiceDiscover(serviceID string) *api.AgentService {

	// 创建Consul客户端连接
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"
	client, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}

	// 获取指定service
	service, _, err := client.Agent().Service(serviceID, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(service.Address)
	fmt.Println(service.Port)

	return service
}

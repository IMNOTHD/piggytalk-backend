package main

import (
	"flag"
	"fmt"
	"os"

	"account/internal/conf"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/contrib/log/fluent/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

const (
	// Name is the name of the compiled software.
	Name = "piggytalk-backend-account"
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server) *kratos.App {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	})
	if err != nil {
		panic(err)
	}

	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
		),
		kratos.Registrar(consul.New(client)),
	)
}

func main() {
	fluentdService := ServiceDiscover("fluentd1")
	logger, err := fluent.NewLogger(
		fmt.Sprintf("tcp://%s:%d", "127.0.0.1", fluentdService.Port),
		fluent.WithTagPrefix(Name))
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

	app, cleanup, err := initApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// ServiceDiscover 服务发现，获取指定id的服务
func ServiceDiscover(serviceID string) *api.AgentService {

	// 创建Consul客户端连接
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := api.NewClient(config)
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

package main

import (
	"flag"
	"fmt"

	"gateway/internal/conf"

	"github.com/go-kratos/kratos/contrib/log/fluent/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id = uuid.New()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id.String()),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	fluentdService := ServiceDiscover("fluentd1")
	fmt.Println(fluentdService.Port)
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

	app, cleanup, err := initApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	_ = logger.Log(log.LevelInfo, Name, fmt.Sprintf("%s is ready to start...", id.String()))

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

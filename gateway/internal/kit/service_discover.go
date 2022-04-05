package kit

import (
	"context"
	"fmt"
	"time"

	consul "github.com/go-kratos/consul/registry"
	kGrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

var ConsulClient *api.Client

const (
	SnowflakeEndpoint = "discovery:///piggytalk-backend-snowflake"
	AccountEndpoint   = "discovery:///piggytalk-backend-account"
	MessageEndpoint   = "discovery:///piggytalk-backend-message"
)

func init() {
	c, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	})
	if err != nil {
		panic(err)
	}
	ConsulClient = c
}

// ServiceDiscover 服务发现，获取指定id的服务
func ServiceDiscover(serviceID string) (*api.AgentService, error) {
	// 获取指定service
	service, _, err := ConsulClient.Agent().Service(serviceID, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(service.Address)
	fmt.Println(service.Port)

	return service, nil
}

func ServiceConn(endpoint string) (*grpc.ClientConn, error) {
	dis := consul.New(ConsulClient)

	return kGrpc.DialInsecure(context.Background(), kGrpc.WithEndpoint(endpoint), kGrpc.WithDiscovery(dis))
}

// ServiceConnWithTimeout 提供timeout选项, 给慢服务使用
func ServiceConnWithTimeout(endpoint string, timeout time.Duration) (*grpc.ClientConn, error) {
	dis := consul.New(ConsulClient)

	return kGrpc.DialInsecure(context.Background(), kGrpc.WithEndpoint(endpoint), kGrpc.WithDiscovery(dis), kGrpc.WithTimeout(timeout))
}

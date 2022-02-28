package kit

import (
	"context"
	"fmt"

	consul "github.com/go-kratos/consul/registry"
	kGrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

var ConsulClient *api.Client

const (
	SnowflakeEndpoint = "discovery:///piggytalk-backend-snowflake"
	AccountEndpoint   = "discovery:///piggytalk-backend-account"
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

const (
	epoch             = int64(1577808000000)                           // 设置起始时间(时间戳/毫秒)：2020-01-01 00:00:00，有效期69年
	timestampBits     = uint(41)                                       // 时间戳占用位数
	datacenteridBits  = uint(3)                                        // 数据中心id所占位数
	workeridBits      = uint(7)                                        // 机器id所占位数
	sequenceBits      = uint(12)                                       // 序列所占的位数
	timestampMax      = int64(-1 ^ (-1 << timestampBits))              // 时间戳最大值
	datacenteridMax   = int64(-1 ^ (-1 << datacenteridBits))           // 支持的最大数据中心id数量
	workeridMax       = int64(-1 ^ (-1 << workeridBits))               // 支持的最大机器id数量
	sequenceMask      = int64(-1 ^ (-1 << sequenceBits))               // 支持的最大序列id数量
	workeridShift     = sequenceBits                                   // 机器id左移位数
	datacenteridShift = sequenceBits + workeridBits                    // 数据中心id左移位数
	timestampShift    = sequenceBits + workeridBits + datacenteridBits // 时间戳左移位数
)

// GetDeviceID 获取数据中心ID和机器ID
func GetDeviceID(sid int64) (dataCenterId, workerId int64) {
	dataCenterId = (sid >> datacenteridShift) & datacenteridMax
	workerId = (sid >> workeridShift) & workeridMax
	return
}

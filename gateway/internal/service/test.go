package service

import (
	"context"

	pb "gateway/api/test"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
)

type TestService struct {
	pb.UnimplementedTestServer

	log *log.Helper
}

func NewTestService(logger log.Logger) *TestService {
	return &TestService{log: log.NewHelper(logger)}
}

func (s *TestService) TestSnowflake(ctx context.Context, req *pb.TestSnowflakeRequest) (*pb.TestSnowflakeReply, error) {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	})

	if err != nil {
		s.log.Error(err.Error())
		panic(err)
	}
	dis := consul.New(client)

	endpoint := "discovery:///piggytalk-backend-snowflake"
	conn, err := grpc.DialInsecure(context.Background(), grpc.WithEndpoint(endpoint), grpc.WithDiscovery(dis))
	if err != nil {
		s.log.Error(err.Error())
		panic(err)
	}
	defer conn.Close()

	c := pb.NewSnowflakeClient(conn)
	r, err := c.CreateSnowflake(ctx, &pb.CreateSnowflakeRequest{
		DataCenterId: req.GetDataCenterId(),
		WorkerId:     req.GetWorkerId(),
	})
	if err != nil {
		s.log.Error(err.Error())
		panic(err)
	}

	return &pb.TestSnowflakeReply{SnowFlakeId: r.GetSnowFlakeId()}, nil
}

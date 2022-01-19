package service

import (
	"context"
	"sync"

	v1 "snowflake/api/snowflake/v1"
	su "snowflake/internal/biz/snowflake/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type Snowflake struct {
	sync.Mutex
	timestamp    int64
	workerId     int64
	dataCenterId int64
	sequence     int64
}

type SnowflakeService struct {
	v1.UnimplementedSnowflakeServer

	su  *su.SnowflakeUsecase
	log *log.Helper
}

func NewSnowflakeService(logger log.Logger) *SnowflakeService {
	return &SnowflakeService{
		log: log.NewHelper(log.With(logger, "module", "account/service/account/v1", "caller", log.DefaultCaller)),
	}
}

func (s *SnowflakeService) CreateSnowflake(ctx context.Context, req *v1.CreateSnowflakeRequest) (*v1.CreateSnowflakeReply, error) {
	s.log.WithContext(ctx).Infof("dataCenterId Received: %d workerId Received: %d", req.GetDataCenterId(), req.GetWorkerId())

	sn, err := s.su.NewSnowflake(req.GetDataCenterId(), req.GetWorkerId())
	if err != nil {
		return nil, err
	}

	return &v1.CreateSnowflakeReply{
		SnowFlakeId: sn.NextVal(),
	}, nil
}

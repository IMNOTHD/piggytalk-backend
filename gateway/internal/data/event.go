package data

import (
	"bytes"
	"context"
	"strconv"

	snV1 "gateway/internal/api/snowflake/snowflake/v1"
	v1 "gateway/internal/biz/event/v1"
	"gateway/internal/conf"
	"gateway/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
)

type eventRepo struct {
	data *Data
	log  *log.Helper
}

func NewEventRepo(data *Data, logger log.Logger) v1.EventRepo {
	return &eventRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "gateway/data/event", "caller", log.DefaultCaller)),
	}
}

func (r *eventRepo) CreateSessionId(ctx context.Context, token string) (v1.SessionId, error) {
	conn, err := kit.ServiceConn(kit.SnowflakeEndpoint)
	if err != nil {
		r.log.Error(err)
		return "", err
	}

	c := snV1.NewSnowflakeClient(conn)
	sr, err := c.CreateSnowflake(ctx, &snV1.CreateSnowflakeRequest{
		DataCenterId: 0,
		WorkerId:     int64(conf.WorkerId),
	})
	if err != nil {
		r.log.Error(err)
		return "", err
	}

	sid := strconv.Itoa(int(sr.GetSnowFlakeId()))

	var buffer bytes.Buffer
	buffer.WriteString("piggytalk:account:sessionId2token:")
	buffer.WriteString(sid)

	x := r.data.Rdb.SAdd(ctx, buffer.String(), token)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	buffer.Reset()
	buffer.WriteString("piggytalk:account:token2sessionId:")
	buffer.WriteString(token)
	x = r.data.Rdb.SAdd(ctx, buffer.String(), sid)
	if x.Err() != nil {
		r.log.Error(x.Err())
		return "", x.Err()
	}

	return v1.SessionId(sid), nil
}

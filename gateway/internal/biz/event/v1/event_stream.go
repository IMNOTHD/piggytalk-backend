package v1

import (
	"context"
	"strconv"
	"time"

	acV1 "gateway/internal/api/account/account/v1"
	snV1 "gateway/internal/api/snowflake/snowflake/v1"
	"gateway/internal/conf"
	"gateway/internal/kit"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type EventRepo interface {
	SendAddFriend(ctx context.Context, sid int64, receiverUuid string, note string, uid string) error
	CreateSessionId(ctx context.Context, token string, sid string, uid string) (SessionId, error)
	RemoveSessionId(ctx context.Context, token string, sid string) error
	SelectToken(ctx context.Context, sessionId string) (string, error)
	SelectUid(ctx context.Context, sessionId string) (string, error)
	UpdateBeatHeart(ctx context.Context, sessionId string, expiration time.Duration) error
	SelectBeatHeart(ctx context.Context, sessionId string) (string, error)
}

type SessionId string

type EventUsecase struct {
	repo EventRepo
	log  *log.Helper
}

func NewEventUsecase(repo EventRepo, logger log.Logger) *EventUsecase {
	return &EventUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "gateway/biz/event/v1", "caller", log.DefaultCaller)),
	}
}

func (uc *EventUsecase) AddFriendRequest(ctx context.Context, receiverUuid uuid.UUID, note string, uid string) (int64, error) {
	conn, err := kit.ServiceConn(kit.SnowflakeEndpoint)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	c := snV1.NewSnowflakeClient(conn)
	sr, err := c.CreateSnowflake(ctx, &snV1.CreateSnowflakeRequest{
		DataCenterId: 0,
		WorkerId:     int64(conf.WorkerId),
	})
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	eid := sr.GetSnowFlakeId()
	err = uc.repo.SendAddFriend(ctx, eid, receiverUuid.String(), note, uid)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	return eid, nil
}

func (uc *EventUsecase) CheckToken(ctx context.Context, token string) (bool, string, error) {
	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		uc.log.Error(err)
		return false, "", err
	}

	c := acV1.NewAccountClient(conn)
	ar, err := c.CheckLoginStat(ctx, &acV1.CheckLoginStatRequest{
		Token: token,
	})
	if err != nil {
		uc.log.Error(err)
		return false, "", err
	}

	if ar.Token != token {
		uc.log.Infof("token %s check failed", token)
		return false, "", nil
	}

	return true, ar.GetUuid(), nil
}

func (uc *EventUsecase) Online(ctx context.Context, token string, uid string) (SessionId, error) {
	conn, err := kit.ServiceConn(kit.SnowflakeEndpoint)
	if err != nil {
		uc.log.Error(err)
		return "", err
	}

	c := snV1.NewSnowflakeClient(conn)
	sr, err := c.CreateSnowflake(ctx, &snV1.CreateSnowflakeRequest{
		DataCenterId: 0,
		WorkerId:     int64(conf.WorkerId),
	})
	if err != nil {
		uc.log.Error(err)
		return "", err
	}

	sid := strconv.Itoa(int(sr.GetSnowFlakeId()))

	s, err := uc.repo.CreateSessionId(ctx, token, sid, uid)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (uc *EventUsecase) BeatHeart(ctx context.Context, sessionId string, expiration time.Duration) error {
	return uc.repo.UpdateBeatHeart(ctx, sessionId, expiration)
}

func (uc *EventUsecase) CheckBeatHeart(ctx context.Context, sessionId string, expiration time.Duration) (bool, error) {
	t, err := uc.repo.SelectBeatHeart(ctx, sessionId)
	if err != nil {
		return false, err
	}
	if t == "" {
		return false, nil
	}

	ts, _ := strconv.ParseInt(t, 10, 64)

	if ts+expiration.Milliseconds() < time.Now().UnixMilli() {
		return false, nil
	}

	return true, nil
}

func (uc *EventUsecase) Offline(ctx context.Context, sessionId string) error {
	token, err := uc.repo.SelectToken(ctx, sessionId)
	if err != nil {
		return err
	}

	return uc.repo.RemoveSessionId(ctx, token, sessionId)
}

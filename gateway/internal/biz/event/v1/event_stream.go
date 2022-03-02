package v1

import (
	"context"
	"strconv"
	"time"

	acV1 "gateway/internal/api/account/account/v1"
	rV1 "gateway/internal/api/account/relation/v1"
	mV1 "gateway/internal/api/message/message/v1"
	snV1 "gateway/internal/api/snowflake/snowflake/v1"
	"gateway/internal/conf"
	"gateway/internal/kit"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type EventRepo interface {
	RabbitMqLister(ctx context.Context) (func(), func())
	SendDeleteFriend(ctx context.Context, uid string, deleteUuid string, sid int64, eventUuid string) error
	SendConfirmFriend(ctx context.Context, sid int64, receiverUuid string, uid string, eventUuid string, addStat string) error
	SendAddFriend(ctx context.Context, sid int64, receiverUuid string, note string, uid string, eventUuid string) error
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

func (uc *EventUsecase) RabbitMqListener(ctx context.Context) (func(), func()) {
	return uc.repo.RabbitMqLister(ctx)
}

func (uc *EventUsecase) ListFriend(ctx context.Context, uid string) ([]string, error) {
	conn, err := kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}

	x := rV1.NewFriendRelationClient(conn)
	r, err := x.ListFriendRelation(ctx, &rV1.ListFriendRelationRequest{UserUUID: uid})
	if err != nil {
		uc.log.Error(err)
		return nil, err
	}

	return r.GetFriendUUID(), nil
}

func (uc *EventUsecase) ConfirmFriendRequest(ctx context.Context, addStat string, eventUuid string) (int64, error) {
	conn, err := kit.ServiceConn(kit.MessageEndpoint)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}
	z := mV1.NewMessageClient(conn)
	mr, err := z.SelectFriendRequest(ctx, &mV1.SelectFriendRequestRequest{EventUuid: eventUuid})
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}
	userAUuid := mr.UserAUuid
	userBUuid := mr.UserBUuid

	conn, err = kit.ServiceConn(kit.SnowflakeEndpoint)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	c := snV1.NewSnowflakeClient(conn)
	sr, err := c.CreateSnowflake(ctx, &snV1.CreateSnowflakeRequest{
		DataCenterId: 0,
		WorkerId:     int64(conf.WorkerId),
	})
	eid := sr.GetSnowFlakeId()

	conn, err = kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	x := rV1.NewFriendRelationClient(conn)
	r, err := x.CreateFriendRelation(ctx, &rV1.CreateFriendRelationRequest{
		UserAUUID: userAUuid,
		UserBUUiD: userBUuid,
	})
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	if !r.Success {
		uc.log.Error("CreateFriend Failed")
		return 0, errors.New(500, "SERVICE_ERROR", "服务错误")
	}

	err = uc.repo.SendConfirmFriend(ctx, eid, userAUuid, userAUuid, eventUuid, addStat)
	if err != nil {
		uc.log.Error(err)
	}

	return eid, nil
}

func (uc *EventUsecase) DeleteFriend(ctx context.Context, deleteUuid string, uid string) (int64, error) {
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

	conn, err = kit.ServiceConn(kit.AccountEndpoint)
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	x := rV1.NewFriendRelationClient(conn)
	r, err := x.DeleteFriendRelation(ctx, &rV1.DeleteFriendRelationRequest{
		UserAUUID: deleteUuid,
		UserBUUiD: uid,
	})
	if err != nil {
		uc.log.Error(err)
		return 0, err
	}

	if !r.Success {
		uc.log.Error("CreateFriend Failed")
		return 0, errors.New(500, "SERVICE_ERROR", "服务错误")
	}

	return eid, nil
}

func (uc *EventUsecase) AddFriendRequest(ctx context.Context, receiverUuid uuid.UUID, note string, uid string, eventUuid string) (int64, error) {
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
	err = uc.repo.SendAddFriend(ctx, eid, receiverUuid.String(), note, uid, eventUuid)
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

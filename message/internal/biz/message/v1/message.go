package v1

import (
	"context"

	pb "message/api/message/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type MessageRepo interface {
	RabbitMqLister(ctx context.Context) (func(), func())
	SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error)
	ListFriendRequest(ctx context.Context, uuid string, startId int64, count int64) ([]*FriendAddMessage, error)
}

type MessageUsecase struct {
	repo MessageRepo
	log  *log.Helper
}

type FriendAddMessage struct {
	EventId      int64
	ReceiverUuid string
	SenderUuid   string
	Type         string
	Ack          string
	EventUuid    string
}

func NewMessageUsecase(repo MessageRepo, logger log.Logger) *MessageUsecase {
	return &MessageUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "message/biz/message/v1", "caller", log.DefaultCaller)),
	}
}

func (uc *MessageUsecase) RabbitMQListener(ctx context.Context) (func(), func()) {
	return uc.repo.RabbitMqLister(ctx)
}

func (uc *MessageUsecase) SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error) {
	return uc.repo.SelectFriendRequest(ctx, eventUuid)
}

func (uc *MessageUsecase) ListFriendRequest(ctx context.Context, uuid string, startId int64, count int64) ([]*pb.ListFriendRequestReply_AddFriendMessage, error) {
	r, err := uc.repo.ListFriendRequest(ctx, uuid, startId, count)
	if err != nil {
		return nil, err
	}

	var k []*pb.ListFriendRequestReply_AddFriendMessage
	for _, message := range r {
		k = append(k, &pb.ListFriendRequestReply_AddFriendMessage{
			EventUuid:    message.EventUuid,
			EventId:      message.EventId,
			Ack:          message.Ack,
			ReceiverUuid: message.ReceiverUuid,
			SenderUuid:   message.SenderUuid,
			Type:         message.Type,
		})
	}

	return k, nil
}

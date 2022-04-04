package v1

import (
	"context"

	pb "message/api/message/v1"
	"message/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type MessageRepo interface {
	RabbitMqLister(ctx context.Context) (func(), func())
	SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error)
	ListFriendRequest(ctx context.Context, uuid string) ([]*data.FriendAddMessage, error)
}

type MessageUsecase struct {
	repo MessageRepo
	log  *log.Helper
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

func (uc *MessageUsecase) ListFriendRequest(ctx context.Context, uuid string) ([]*pb.ListFriendRequestReply_AddFriendMessage, error) {
	r, err := uc.repo.ListFriendRequest(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var k []*pb.ListFriendRequestReply_AddFriendMessage
	for _, message := range r {
		k = append(k, &pb.ListFriendRequestReply_AddFriendMessage{
			EventUuid:    message.EventUuid,
			EventId:      message.EventId,
			Ack:          message.Ack,
			ReceiverUuid: message.UserB.String(),
			SenderUuid:   message.UserA.String(),
			Type:         message.Type,
		})
	}

	return k, nil
}

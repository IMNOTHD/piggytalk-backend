package v1

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type MessageRepo interface {
	RabbitMqLister(ctx context.Context) (func(), func())
	SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error)
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

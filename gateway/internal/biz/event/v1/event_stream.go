package v1

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

type EventRepo interface {
	CreateSessionId(ctx context.Context, token string) (SessionId, error)
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

func (uc *EventUsecase) Online(ctx context.Context, token string) (SessionId, error) {
	sid, err := uc.repo.CreateSessionId(ctx, token)
	if err != nil {
		return "", err
	}

	return sid, nil
}

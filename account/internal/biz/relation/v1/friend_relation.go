package v1

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type FriendRelationRepo interface {
	CreateFriend(ctx context.Context, a uuid.UUID, b uuid.UUID) error
	DeleteFriend(ctx context.Context, a uuid.UUID, b uuid.UUID) error
	ListFriend(ctx context.Context, a uuid.UUID) ([]string, error)
}

type FriendRelationUsecase struct {
	repo FriendRelationRepo
	log  *log.Helper
}

func NewFriendRelationUsecase(repo FriendRelationRepo, logger log.Logger) *FriendRelationUsecase {
	return &FriendRelationUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "account/biz/relation/v1/friend_relation", "caller", log.DefaultCaller)),
	}
}

func (uc *FriendRelationUsecase) CreateFriend(ctx context.Context, userA string, userB string) error {
	return uc.repo.CreateFriend(ctx, uuid.MustParse(userA), uuid.MustParse(userB))
}

func (uc *FriendRelationUsecase) RemoveFriend(ctx context.Context, userA string, userB string) error {
	return uc.repo.DeleteFriend(ctx, uuid.MustParse(userA), uuid.MustParse(userB))
}

func (uc *FriendRelationUsecase) ListFriend(ctx context.Context, user string) ([]string, error) {
	return uc.repo.ListFriend(ctx, uuid.MustParse(user))
}

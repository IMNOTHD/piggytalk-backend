package service

import (
	"context"

	pb "account/api/relation/v1"
	v1 "account/internal/biz/relation/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type FriendRelationService struct {
	pb.UnimplementedFriendRelationServer

	fu  *v1.FriendRelationUsecase
	log *log.Helper
}

// NewFriendRelationService 说明如下: 这里仅作为添加关系, 添加请求使用message, 仅在确认后使用这些
func NewFriendRelationService(fu *v1.FriendRelationUsecase, logger log.Logger) *FriendRelationService {
	return &FriendRelationService{
		fu:  fu,
		log: log.NewHelper(log.With(logger, "module", "account/service/relation/v1/friend_relation", "caller", log.DefaultCaller)),
	}
}

func (s *FriendRelationService) CreateFriendRelation(ctx context.Context, req *pb.CreateFriendRelationRequest) (*pb.CreateFriendRelationReply, error) {
	err := s.fu.CreateFriend(ctx, req.GetUserAUUID(), req.GetUserBUUiD())
	if err != nil {
		return nil, err
	}

	return &pb.CreateFriendRelationReply{Success: true}, nil
}
func (s *FriendRelationService) DeleteFriendRelation(ctx context.Context, req *pb.DeleteFriendRelationRequest) (*pb.DeleteFriendRelationReply, error) {
	err := s.fu.RemoveFriend(ctx, req.GetUserAUUID(), req.GetUserBUUiD())
	if err != nil {
		return nil, err
	}

	return &pb.DeleteFriendRelationReply{Success: true}, nil
}
func (s *FriendRelationService) ListFriendRelation(ctx context.Context, req *pb.ListFriendRelationRequest) (*pb.ListFriendRelationReply, error) {
	r, err := s.fu.ListFriend(ctx, req.GetUserUUID())
	if err != nil {
		return nil, err
	}

	return &pb.ListFriendRelationReply{FriendUUID: r}, nil
}

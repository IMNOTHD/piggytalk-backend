package service

import (
	"context"

	pb "message/api/message/v1"
	v1 "message/internal/biz/message/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type MessageService struct {
	pb.UnimplementedMessageServer

	eu  *v1.MessageUsecase
	log *log.Helper
}

func NewMessageService(eu *v1.MessageUsecase, logger log.Logger) *MessageService {
	service := &MessageService{
		eu:  eu,
		log: log.NewHelper(log.With(logger, "module", "message/service/message/v1", "caller", log.DefaultCaller)),
	}

	go service.rabbitmqListener()

	return service
}

func (s *MessageService) rabbitmqListener() {
	m, e := s.eu.RabbitMQListener(context.Background())
	go m()
	go e()
}

func (s *MessageService) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.CreateMessageReply, error) {
	return &pb.CreateMessageReply{}, nil
}

func (s *MessageService) SelectFriendRequest(ctx context.Context, req *pb.SelectFriendRequestRequest) (*pb.SelectFriendRequestReply, error) {
	a, b, err := s.eu.SelectFriendRequest(ctx, req.GetEventUuid())
	if err != nil {
		return nil, err
	}

	return &pb.SelectFriendRequestReply{
		UserAUuid: a,
		UserBUuid: b,
	}, nil
}

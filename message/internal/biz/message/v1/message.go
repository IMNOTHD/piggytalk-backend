package v1

import (
	"context"

	pb "message/api/message/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang/protobuf/proto"
)

type MessageRepo interface {
	RabbitMqLister(ctx context.Context) (func(), func())
	SelectFriendRequest(ctx context.Context, eventUuid string) (string, string, error)
	ListFriendRequest(ctx context.Context, uuid string, startId int64, count int64) ([]*FriendAddMessage, error)
	ListSingleMessage(ctx context.Context, uid string, friendUuid string, startId int64, count int64) ([]*SingleMessage, error)
	ListUnAckSingleMessage(ctx context.Context, uid string) ([]*UnAckMessage, error)
}

type MessageUsecase struct {
	repo MessageRepo
	log  *log.Helper
}

type UnAckMessage struct {
	UnAck      int64
	FriendUuid string
}

type FriendAddMessage struct {
	EventId      int64
	ReceiverUuid string
	SenderUuid   string
	Type         string
	Ack          string
	EventUuid    string
}

type SingleMessage struct {
	MessageId   int64
	MessageUuid string
	AlreadyRead bool
	SenderUuid  string
	Talk        string
	Message     []byte
}

func NewMessageUsecase(repo MessageRepo, logger log.Logger) *MessageUsecase {
	return &MessageUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "message/biz/message/v1", "caller", log.DefaultCaller)),
	}
}

func (uc *MessageUsecase) ListUnAckSingleMessage(ctx context.Context, uuid string) ([]*pb.ListUnAckSingleMessageResponse_UnackSingleMessage, error) {
	x, err := uc.repo.ListUnAckSingleMessage(ctx, uuid)
	if err != nil {
		return nil, err
	}

	var k []*pb.ListUnAckSingleMessageResponse_UnackSingleMessage
	for _, message := range x {
		k = append(k, &pb.ListUnAckSingleMessageResponse_UnackSingleMessage{
			FriendUuid: message.FriendUuid,
			UnAck:      message.UnAck,
		})
	}

	return k, nil
}

func (uc *MessageUsecase) ListSingleMessage(ctx context.Context, uuid string, friendUuid string, startId int64, count int64) ([]*pb.ListSingleMessageResponse_MessageStruct, error) {
	x, err := uc.repo.ListSingleMessage(ctx, uuid, friendUuid, startId, count)
	if err != nil {
		return nil, err
	}

	var messageStruct []*pb.ListSingleMessageResponse_MessageStruct
	for _, message := range x {
		// todo protobuf blob
		var s *pb.ListSingleMessageResponse_MessageStruct_SingleMessage
		err := proto.Unmarshal(message.Message, s)

		if err != nil {
			uc.log.Warn(err)
			continue
		}

		messageStruct = append(messageStruct, &pb.ListSingleMessageResponse_MessageStruct{
			SingleMessage: s,
			MessageId:     message.MessageId,
			MessageUuid:   message.MessageUuid,
			SenderUuid:    message.SenderUuid,
		})
	}

	return messageStruct, nil
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

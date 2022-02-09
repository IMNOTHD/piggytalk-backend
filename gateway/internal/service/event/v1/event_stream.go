package v1

import (
	"context"
	"io"

	pb "gateway/api/event/v1"
	v1 "gateway/internal/biz/event/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type EventStreamService struct {
	pb.UnimplementedEventStreamServer

	eu  *v1.EventUsecase
	log *log.Helper
}

func NewEventStreamService(eu *v1.EventUsecase, logger log.Logger) *EventStreamService {
	return &EventStreamService{
		eu:  eu,
		log: log.NewHelper(log.With(logger, "module", "gateway/service/event/v1", "caller", log.DefaultCaller)),
	}
}

func (s *EventStreamService) EventStream(conn pb.EventStream_EventStreamServer) error {
	var ctx context.Context
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			s.log.Infof("token %s end stream...", req.GetToken())
			return nil
		}
		if err != nil {
			s.log.Error(err)
			return err
		}

		switch req.Event.(type) {
		case *pb.EventStreamRequest_OnlineRequest:
			_, err := s.eu.Online(ctx, req.GetOnlineRequest().GetToken())
			if err != nil {

			}
		case *pb.EventStreamRequest_BeatHeartRequest:

		case *pb.EventStreamRequest_OfflineRequest:

		}

		//err = conn.Send(&pb.EventStreamResponse{})
		//if err != nil {
		//	return err
		//}
	}
}

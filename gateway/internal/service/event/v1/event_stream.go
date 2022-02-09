package v1

import (
	"io"

	pb "gateway/api/event/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type EventStreamService struct {
	pb.UnimplementedEventStreamServer

	log *log.Helper
}

func NewEventStreamService(logger log.Logger) *EventStreamService {
	return &EventStreamService{
		log: log.NewHelper(log.With(logger, "module", "gateway/service/event/v1", "caller", log.DefaultCaller)),
	}
}

func (s *EventStreamService) EventStream(conn pb.EventStream_EventStreamServer) error {
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

			break
		case *pb.EventStreamRequest_BeatHeartRequest:
			break
		case *pb.EventStreamRequest_OfflineRequest:
			break
		}

		//err = conn.Send(&pb.EventStreamResponse{})
		//if err != nil {
		//	return err
		//}
	}
}

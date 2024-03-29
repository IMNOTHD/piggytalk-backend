package v1

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"sync"
	"time"

	pb "gateway/api/event/v1"
	v1 "gateway/internal/biz/event/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type EventStreamService struct {
	pb.UnimplementedEventStreamServer

	eu  *v1.EventUsecase
	log *log.Helper
}

type lastMessage struct {
	MessageUuid uuid.UUID
	SendTime    int64
}
type lastEvent struct {
	EventUuid uuid.UUID
	SendTime  int64
}

type Message struct {
	Type      string
	Body      string
	MessageId string
}

// SessionReceiveMq 参数为uid
type SessionReceiveMq map[string]chan Message

// message type
const ()

// event type
const (
	_addFriend    = "AddFriend"
	_createFriend = "CreateFriend"
)

var (
	ReceiveMessageMq SessionReceiveMq
	ReceiveEventMq   SessionReceiveMq
)

func NewEventStreamService(eu *v1.EventUsecase, logger log.Logger) *EventStreamService {
	ReceiveEventMq = make(SessionReceiveMq)
	ReceiveMessageMq = make(SessionReceiveMq)

	service := &EventStreamService{
		eu:  eu,
		log: log.NewHelper(log.With(logger, "module", "gateway/service/event/v1", "caller", log.DefaultCaller)),
	}

	go service.rabbitmqListener()

	return service
}

const (
	// 允许channel缓冲的长度, 一个连接一个channel, 没必要开的很大, 消费者同理
	_commodityLoad = 64
	// 消费者数量
	_consumerNumber = 4
	// 心跳检测间隔
	_maxBeatHeartDuration = time.Second * 45
	// 心跳有效期
	_beatHeartExpiration = time.Second * 60
)

func (s *EventStreamService) rabbitmqListener() {
	m, e := s.eu.RabbitMqListener(context.Background())
	go m()
	go e()
}

func (s *EventStreamService) EventStream(conn pb.EventStream_EventStreamServer) error {
	var wg sync.WaitGroup
	var e error
	var exit = false
	var sessionId string = ""
	uid := ""
	ctx := context.Background()
	token := ""
	//lm := lastMessage{
	//	MessageUuid: uuid.UUID{},
	//	SendTime:    0,
	//}
	le := lastEvent{
		EventUuid: uuid.UUID{},
		SendTime:  0,
	}

	ch := make(chan pb.EventStreamRequest, _commodityLoad)
	beatStartCh := make(chan bool)
	stopCh := make(chan struct{})

	consumer := func(sc <-chan struct{}, c chan pb.EventStreamRequest) {
		defer wg.Done()

		go func() {
			// 登录成功后, 可以接收消息
		ForEnd:
			for {
				select {
				case <-sc:
					return
				default:
					if uid != "" {
						break ForEnd
					}
				}
			}

			if ReceiveEventMq[uid] != nil {
				ReceiveEventMq[uid] = make(chan Message, _commodityLoad)
			}
			if ReceiveMessageMq[uid] != nil {
				ReceiveMessageMq[uid] = make(chan Message, _commodityLoad)
			}

			for {
				select {
				case <-sc:
					return
				case r, ok := <-ReceiveEventMq[uid]:
					if !ok {
						return
					}

					switch r.Type {
					case _addFriend:
						type body struct {
							ReceiverUuid string
							Note         string
							Uid          string
							EventUuid    string
						}
						var b body
						err := json.Unmarshal([]byte(r.Body), &b)
						if err != nil {
							s.log.Error(err)
							continue
						}

						m, err := strconv.ParseInt(r.MessageId, 10, 64)
						if err != nil {
							s.log.Error(err)
							continue
						}

						err = conn.Send(&pb.EventStreamResponse{
							Token:    token,
							Code:     pb.Code_OK,
							Messages: "",
							Event: &pb.EventStreamResponse_NotifyReceiveAddFriendResponse{
								NotifyReceiveAddFriendResponse: &pb.NotifyReceiveAddFriendResponse{
									EventId:     m,
									RequestUuid: b.Uid,
									Note:        b.Note,
								},
							},
						})
						if err != nil {
							s.log.Error(err)
							continue
						}
					}
				case r, ok := <-ReceiveMessageMq[uid]:
					if !ok {
						return
					}
					switch r.Type {

					}
				}
			}
		}()

		for {
			if exit {
				return
			}

			select {
			// 是否关闭
			case <-sc:
				return
			// 处理客户端发出消息
			case req, ok := <-c:
				if !ok {
					return
				}

				switch req.Event.(type) {
				case *pb.EventStreamRequest_OnlineRequest:
					sid, err := s.eu.Online(ctx, req.GetOnlineRequest().GetToken(), uid)
					if err != nil {
						e = err
						err := conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_ABORTED,
							Messages: "认证服务错误",
						})
						if err != nil {
							s.log.Error(err)
						}
						// 认证出错, 该连接无存在必要, 断开
						exit = true
						break
					}

					token = req.GetToken()

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_OnlineResponse{
							OnlineResponse: &pb.OnlineResponse{SessionId: string(sid)},
						},
					})
					if err != nil {
						s.log.Error(err)
						e = err
						exit = true
					}

					sessionId = string(sid)

					// 登录即视为一次心跳
					err = s.eu.BeatHeart(ctx, string(sid), _beatHeartExpiration)
					if err != nil {
						s.log.Error(err)
					}

					// 开始计算心跳
					beatStartCh <- true
				case *pb.EventStreamRequest_BeatHeartRequest:
					if uid == "" {
						continue
					}

					err := s.eu.BeatHeart(ctx, req.GetBeatHeartRequest().GetSessionId(), _beatHeartExpiration)
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "心跳错误",
							Event: &pb.EventStreamResponse_BeatHeartResponse{
								BeatHeartResponse: &pb.BeatHeartResponse{Flag: pb.BeatHeartResponse_FIN},
							},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_BeatHeartResponse{
							BeatHeartResponse: &pb.BeatHeartResponse{Flag: pb.BeatHeartResponse_ACK},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_OfflineRequest:
					if uid == "" {
						continue
					}

					err := s.eu.Offline(ctx, sessionId)
					if err != nil {
						s.log.Error(err)
					}

					exit = true
				case *pb.EventStreamRequest_AddFriendRequest:
					if uid == "" {
						continue
					}

					k := lastEvent{EventUuid: uuid.MustParse(req.GetAddFriendRequest().GetEventUuid()), SendTime: req.GetAddFriendRequest().GetSendTime()}
					if k == le {
						continue
					}

					eid, err := s.eu.AddFriendRequest(ctx, uuid.MustParse(req.GetAddFriendRequest().GetReceiverUuid()), req.GetAddFriendRequest().GetNote(), uid, req.GetAddFriendRequest().GetEventUuid())
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event: &pb.EventStreamResponse_AddFriendResponse{
								AddFriendResponse: &pb.AddFriendResponse{EventUuid: req.GetAddFriendRequest().GetEventUuid()},
							},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_AddFriendResponse{
							AddFriendResponse: &pb.AddFriendResponse{
								EventUuid: req.GetAddFriendRequest().GetEventUuid(),
								EventId:   eid,
							},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_ConfirmFriendRequest:
					if uid == "" {
						continue
					}

					k := lastEvent{EventUuid: uuid.MustParse(req.GetConfirmFriendRequest().GetEventUuid()), SendTime: req.GetConfirmFriendRequest().GetSendTime()}
					if k == le {
						continue
					}

					st := "WAITING"
					switch req.GetConfirmFriendRequest().GetAddStatCode() {
					case pb.AddStatCode_SUCCESS:
						st = "SUCCESS"
					case pb.AddStatCode_DENIED:
						st = "DENIED"
					}

					eid, err := s.eu.ConfirmFriendRequest(ctx, st, req.GetConfirmFriendRequest().GetEventUuid())
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event: &pb.EventStreamResponse_ConfirmFriendResponse{
								ConfirmFriendResponse: &pb.ConfirmFriendResponse{EventUuid: req.GetConfirmFriendRequest().GetEventUuid()},
							},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_ConfirmFriendResponse{
							ConfirmFriendResponse: &pb.ConfirmFriendResponse{
								EventUuid: req.GetAddFriendRequest().GetEventUuid(),
								EventId:   eid,
							},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_ListFriendRequest:
					if uid == "" {
						continue
					}

					l, err := s.eu.ListFriend(ctx, uid)
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event:    &pb.EventStreamResponse_ListFriendResponse{},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_ListFriendResponse{
							ListFriendResponse: &pb.ListFriendResponse{FriendUuid: l},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_DeleteFriendRequest:
					if uid == "" {
						continue
					}

					k := lastEvent{EventUuid: uuid.MustParse(req.GetDeleteFriendRequest().GetEventUuid()), SendTime: req.GetDeleteFriendRequest().GetSendTime()}
					if k == le {
						continue
					}

					eid, err := s.eu.DeleteFriend(ctx, uid, req.GetDeleteFriendRequest().GetRemoveUuid())
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event: &pb.EventStreamResponse_DeleteFriendResponse{
								DeleteFriendResponse: &pb.DeleteFriendResponse{EventUuid: req.GetConfirmFriendRequest().GetEventUuid()},
							},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_DeleteFriendResponse{
							DeleteFriendResponse: &pb.DeleteFriendResponse{
								EventUuid: req.GetAddFriendRequest().GetEventUuid(),
								EventId:   eid,
							},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_ListUserInfoRequest:
					if uid == "" {
						continue
					}

					l, err := s.eu.ListUserInfo(ctx, req.GetListUserInfoRequest().GetUuid())
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event:    &pb.EventStreamResponse_ListUserInfoResponse{},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_ListUserInfoResponse{
							ListUserInfoResponse: &pb.ListUserInfoResponse{Userinfo: l},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_ListFriendRequestRequest:
					if uid == "" {
						continue
					}

					r, err := s.eu.ListFriendRequest(ctx, uid, req.GetListFriendRequestRequest().GetStartEventId(), req.GetListFriendRequestRequest().GetCount())
					if err != nil {
						s.log.Error(err)
						err = conn.Send(&pb.EventStreamResponse{
							Token:    req.GetToken(),
							Code:     pb.Code_UNAVAILABLE,
							Messages: "服务错误",
							Event:    &pb.EventStreamResponse_ListFriendRequestResponse{},
						})
						if err != nil {
							s.log.Error(err)
						}
					}

					err = conn.Send(&pb.EventStreamResponse{
						Token:    req.GetToken(),
						Code:     pb.Code_OK,
						Messages: "",
						Event: &pb.EventStreamResponse_ListFriendRequestResponse{
							ListFriendRequestResponse: &pb.ListFriendRequestResponse{AddFriendMessage: r},
						},
					})
					if err != nil {
						s.log.Error(err)
					}
				case *pb.EventStreamRequest_AckFriendMessageRequest:
					// 不需要可靠, 发出即可
					if uid == "" {
						continue
					}

					_ = s.eu.AckFriendMessage(ctx, uid, req.GetAckFriendMessageRequest().GetEventId())
				case *pb.EventStreamRequest_ListUnAckSingleMessageRequest:
				}
			}

		}
	}

	for i := 0; i < _consumerNumber; i++ {
		go consumer(stopCh, ch)
		wg.Add(1)
	}

	// 心跳检测
	go func() {
		select {
		case <-stopCh:
			return
		case _, ok := <-beatStartCh:
			if !ok {
				return
			}
		}

		for {
			time.Sleep(_maxBeatHeartDuration)
			select {
			case <-stopCh:
				return
			default:
				b, err := s.eu.CheckBeatHeart(ctx, sessionId, _beatHeartExpiration)
				if err != nil {
					s.log.Error(err)
				}
				// 心跳过期, 切断连接
				if !b {
					s.log.Infof("sessionId %s beatheart expired, disconnect", sessionId)
					err := s.eu.Offline(ctx, sessionId)
					if err != nil {
						s.log.Error(err)
					}
					exit = true
					return
				}
			}
		}
	}()

	for {
		if exit {
			close(stopCh)
			break
		}

		req, err := conn.Recv()
		if err == io.EOF {
			s.log.Infof("token %s end stream...", req.GetToken())
			exit = true
		}
		if err != nil {
			s.log.Error(err)
			e = err
			exit = true
		}
		if req == nil {
			continue
		}

		s.log.Infof("uid: %s, req: %s", uid, req.String())

		f, u, err := s.eu.CheckToken(ctx, req.GetToken())
		if err != nil {
			s.log.Error(err)
			e = err
			exit = true
		}
		if f == false {
			err = conn.Send(&pb.EventStreamResponse{
				Token:    req.GetToken(),
				Code:     pb.Code_UNAUTHENTICATED,
				Messages: "登录失效，请重新登录",
				Event: &pb.EventStreamResponse_OfflineResponse{
					OfflineResponse: &pb.OfflineResponse{Token: req.GetToken()}},
			})
			if err != nil {
				s.log.Error(err)
				e = err
			}
			exit = true
			continue
		}
		uid = u

		ch <- *req
	}

	close(beatStartCh)
	// 等待通道全部关闭
	for len(ch) != 0 {
	}
	close(ch)
	wg.Wait()

	return e
}

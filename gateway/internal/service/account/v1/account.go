package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "gateway/api/account/v1"
	"gateway/internal/biz/account/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/valyala/fasthttp"
)

type AccountService struct {
	pb.UnimplementedAccountServer

	au  *v1.AccountUsecase
	log *log.Helper
}

func NewAccountService(au *v1.AccountUsecase, logger log.Logger) *AccountService {
	return &AccountService{
		au:  au,
		log: log.NewHelper(log.With(logger, "module", "gateway/service/account/v1", "caller", log.DefaultCaller))}
}

func (s *AccountService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	err := s.verifyReCaptchaToken(req.ReCaptchaToken)
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}

	a, t, err := s.au.Login(ctx, &v1.Account{
		Username: req.GetAccount(),
		Password: req.GetPassword(),
	})
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}

	if tr, ok := transport.FromServerContext(ctx); ok {
		tr.ReplyHeader().Set("Set-Cookie", fmt.Sprintf("csrf_token=%s; Path=/; Expires=%s; HttpOnly", string(t), time.Now().AddDate(1, 0, 0).Format(http.TimeFormat)))
	}

	return &pb.LoginReply{
		Token:    string(t),
		Username: a.Username,
		Email:    a.Email,
		Phone:    a.Phone,
		Avatar:   a.Avatar,
		Nickname: a.Nickname,
	}, nil
}
func (s *AccountService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	err := s.verifyReCaptchaToken(req.ReCaptchaToken)
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}

	t, err := s.au.Register(ctx, &v1.Account{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Nickname: req.GetNickname(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Avatar:   req.GetAvatar(),
	})
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}

	if tr, ok := transport.FromServerContext(ctx); ok {
		tr.ReplyHeader().Set("Set-Cookie", fmt.Sprintf("csrf_token=%s; Path=/; Expires=%s; HttpOnly", string(t), time.Now().AddDate(1, 0, 0).Format(http.TimeFormat)))
	}

	return &pb.RegisterReply{Token: string(t)}, nil
}

func (s *AccountService) UpdateAvatar(ctx context.Context, req *pb.UpdateAvatarRequest) (*pb.UpdateAvatarReply, error) {
	err := s.au.UpdateAvatar(ctx, req.GetToken(), req.GetAvatar())
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}

	return &pb.UpdateAvatarReply{
		Token:  req.GetToken(),
		Avatar: req.GetAvatar(),
	}, nil
}

func (s *AccountService) SearchUuid(ctx context.Context, req *pb.SearchUuidRequest) (*pb.SearchUuidReply, error) {
	r, err := s.au.SearchUuid(ctx, req.GetUuid())
	if err != nil {
		s.log.Error(err)
		return nil, errorCheck(err)
	}
	if r != nil {
		return &pb.SearchUuidReply{
			Uuid:     req.GetUuid(),
			Avatar:   r.Avatar,
			Nickname: r.Nickname,
		}, nil
	}

	return nil, nil
}

// 拦截错误, 防止把内部错误带出去
func errorCheck(err error) error {
	e := errors.FromError(err)
	if e.GetCode() < 400 || e.Code >= 500 {
		return errors.New(500, "SERVICE_ERROR", "服务错误")
	}
	return err
}

type RecaptchaResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

const _accessScore = 0.7

func (s *AccountService) verifyReCaptchaToken(token string) error {
	url := `https://www.recaptcha.net/recaptcha/api/siteverify`

	args := &fasthttp.Args{}
	args.Add("secret", "6LfCMxQeAAAAADjayvh5ui3vfihYx2TSpVE_cPuH")
	args.Add("response", token)

	st, r, err := fasthttp.Post(nil, url, args)
	if err != nil {
		s.log.Error(err)
		return err
	}
	if st != fasthttp.StatusOK {
		s.log.Error(fmt.Sprintf("verify service error: %v %v", st, r))
		return errors.New(500, "SERVICE_ERROR", "SERVICE_ERROR")
	}

	var re RecaptchaResponse
	err = json.Unmarshal(r, &re)
	if err != nil {
		s.log.Error(err)
		return err
	}
	if re.Success == false {
		s.log.Error(fmt.Sprintf("verify service error: %v", re))
		return errors.New(500, "SERVICE_ERROR", "SERVICE_ERROR")
	}

	if re.Score < _accessScore {
		s.log.Infof(fmt.Sprintf("verify deniend: %s", token))
		return errors.New(403, "FORBIDDEN", "FORBIDDEN")
	}

	return nil
}

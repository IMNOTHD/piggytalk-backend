package server

import (
	httpNet "net/http"

	v12 "gateway/api/account/v1"
	v2 "gateway/api/test"
	"gateway/internal/conf"
	"gateway/internal/service"
	v1 "gateway/internal/service/account/v1"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, test *service.TestService, account *v1.AccountService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
		http.Filter(handlers.CORS(
			handlers.AllowedOrigins([]string{"http://localhost:3000"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowCredentials(),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	opts = append(opts, http.ResponseEncoder(func(w httpNet.ResponseWriter, r *httpNet.Request, i interface{}) error {
		type response struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}
		reply := &response{
			Code:    200,
			Message: "ok",
			Data:    i,
		}
		codec := encoding.GetCodec("json")
		data, err := codec.Marshal(reply)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return nil
	}))

	srv := http.NewServer(opts...)

	v2.RegisterTestHTTPServer(srv, test)
	v12.RegisterAccountHTTPServer(srv, account)
	return srv
}

package server

import (
	helloWorldV1 "snowflake/api/helloworld/v1"
	snowflakeV1 "snowflake/api/snowflake/v1"
	"snowflake/internal/conf"
	"snowflake/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, snowflake *service.SnowflakeService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
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
	srv := http.NewServer(opts...)
	helloWorldV1.RegisterGreeterHTTPServer(srv, greeter)
	snowflakeV1.RegisterSnowflakeHTTPServer(srv, snowflake)
	return srv
}

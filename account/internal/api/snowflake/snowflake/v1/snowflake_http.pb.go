// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http v2.1.3

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

type SnowflakeHTTPServer interface {
	CreateSnowflake(context.Context, *CreateSnowflakeRequest) (*CreateSnowflakeReply, error)
}

func RegisterSnowflakeHTTPServer(s *http.Server, srv SnowflakeHTTPServer) {
	r := s.Route("/")
	r.GET("/snowflake", _Snowflake_CreateSnowflake0_HTTP_Handler(srv))
}

func _Snowflake_CreateSnowflake0_HTTP_Handler(srv SnowflakeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in CreateSnowflakeRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/snowflake.api.snowflake.v1.Snowflake/CreateSnowflake")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.CreateSnowflake(ctx, req.(*CreateSnowflakeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*CreateSnowflakeReply)
		return ctx.Result(200, reply)
	}
}

type SnowflakeHTTPClient interface {
	CreateSnowflake(ctx context.Context, req *CreateSnowflakeRequest, opts ...http.CallOption) (rsp *CreateSnowflakeReply, err error)
}

type SnowflakeHTTPClientImpl struct {
	cc *http.Client
}

func NewSnowflakeHTTPClient(client *http.Client) SnowflakeHTTPClient {
	return &SnowflakeHTTPClientImpl{client}
}

func (c *SnowflakeHTTPClientImpl) CreateSnowflake(ctx context.Context, in *CreateSnowflakeRequest, opts ...http.CallOption) (*CreateSnowflakeReply, error) {
	var out CreateSnowflakeReply
	pattern := "/snowflake"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation("/snowflake.api.snowflake.v1.Snowflake/CreateSnowflake"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
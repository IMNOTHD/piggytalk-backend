// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.1
// source: api/test/test.proto

package test

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TestClient is the client API for Test service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TestClient interface {
	TestSnowflake(ctx context.Context, in *TestSnowflakeRequest, opts ...grpc.CallOption) (*TestSnowflakeReply, error)
}

type testClient struct {
	cc grpc.ClientConnInterface
}

func NewTestClient(cc grpc.ClientConnInterface) TestClient {
	return &testClient{cc}
}

func (c *testClient) TestSnowflake(ctx context.Context, in *TestSnowflakeRequest, opts ...grpc.CallOption) (*TestSnowflakeReply, error) {
	out := new(TestSnowflakeReply)
	err := c.cc.Invoke(ctx, "/api.test.Test/TestSnowflake", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TestServer is the server API for Test service.
// All implementations must embed UnimplementedTestServer
// for forward compatibility
type TestServer interface {
	TestSnowflake(context.Context, *TestSnowflakeRequest) (*TestSnowflakeReply, error)
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestServer must be embedded to have forward compatible implementations.
type UnimplementedTestServer struct {
}

func (UnimplementedTestServer) TestSnowflake(context.Context, *TestSnowflakeRequest) (*TestSnowflakeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestSnowflake not implemented")
}
func (UnimplementedTestServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestServer will
// result in compilation errors.
type UnsafeTestServer interface {
	mustEmbedUnimplementedTestServer()
}

func RegisterTestServer(s grpc.ServiceRegistrar, srv TestServer) {
	s.RegisterService(&Test_ServiceDesc, srv)
}

func _Test_TestSnowflake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestSnowflakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).TestSnowflake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.test.Test/TestSnowflake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).TestSnowflake(ctx, req.(*TestSnowflakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Test_ServiceDesc is the grpc.ServiceDesc for Test service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Test_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.test.Test",
	HandlerType: (*TestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TestSnowflake",
			Handler:    _Test_TestSnowflake_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/test/test.proto",
}

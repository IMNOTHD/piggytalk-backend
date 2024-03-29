// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.1
// source: api/upload/v1/upload.proto

package v1

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

// UploadClient is the client API for Upload service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UploadClient interface {
	ImgUpload(ctx context.Context, opts ...grpc.CallOption) (Upload_ImgUploadClient, error)
}

type uploadClient struct {
	cc grpc.ClientConnInterface
}

func NewUploadClient(cc grpc.ClientConnInterface) UploadClient {
	return &uploadClient{cc}
}

func (c *uploadClient) ImgUpload(ctx context.Context, opts ...grpc.CallOption) (Upload_ImgUploadClient, error) {
	stream, err := c.cc.NewStream(ctx, &Upload_ServiceDesc.Streams[0], "/gateway.api.upload.v1.Upload/ImgUpload", opts...)
	if err != nil {
		return nil, err
	}
	x := &uploadImgUploadClient{stream}
	return x, nil
}

type Upload_ImgUploadClient interface {
	Send(*ImgUploadRequest) error
	CloseAndRecv() (*ImgUploadResponse, error)
	grpc.ClientStream
}

type uploadImgUploadClient struct {
	grpc.ClientStream
}

func (x *uploadImgUploadClient) Send(m *ImgUploadRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *uploadImgUploadClient) CloseAndRecv() (*ImgUploadResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ImgUploadResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UploadServer is the server API for Upload service.
// All implementations must embed UnimplementedUploadServer
// for forward compatibility
type UploadServer interface {
	ImgUpload(Upload_ImgUploadServer) error
	mustEmbedUnimplementedUploadServer()
}

// UnimplementedUploadServer must be embedded to have forward compatible implementations.
type UnimplementedUploadServer struct {
}

func (UnimplementedUploadServer) ImgUpload(Upload_ImgUploadServer) error {
	return status.Errorf(codes.Unimplemented, "method ImgUpload not implemented")
}
func (UnimplementedUploadServer) mustEmbedUnimplementedUploadServer() {}

// UnsafeUploadServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UploadServer will
// result in compilation errors.
type UnsafeUploadServer interface {
	mustEmbedUnimplementedUploadServer()
}

func RegisterUploadServer(s grpc.ServiceRegistrar, srv UploadServer) {
	s.RegisterService(&Upload_ServiceDesc, srv)
}

func _Upload_ImgUpload_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(UploadServer).ImgUpload(&uploadImgUploadServer{stream})
}

type Upload_ImgUploadServer interface {
	SendAndClose(*ImgUploadResponse) error
	Recv() (*ImgUploadRequest, error)
	grpc.ServerStream
}

type uploadImgUploadServer struct {
	grpc.ServerStream
}

func (x *uploadImgUploadServer) SendAndClose(m *ImgUploadResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *uploadImgUploadServer) Recv() (*ImgUploadRequest, error) {
	m := new(ImgUploadRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Upload_ServiceDesc is the grpc.ServiceDesc for Upload service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Upload_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gateway.api.upload.v1.Upload",
	HandlerType: (*UploadServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ImgUpload",
			Handler:       _Upload_ImgUpload_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "api/upload/v1/upload.proto",
}

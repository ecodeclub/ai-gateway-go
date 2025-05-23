// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: ai/v1/ai.proto

package aiv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AIService_Invoke_FullMethodName = "/ai.v1.AIService/Invoke"
	AIService_Stream_FullMethodName = "/ai.v1.AIService/Stream"
)

// AIServiceClient is the client API for AIService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AIServiceClient interface {
	Invoke(ctx context.Context, in *LLMRequest, opts ...grpc.CallOption) (*LLMResponse, error)
	Stream(ctx context.Context, in *LLMRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamEvent], error)
}

type aIServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAIServiceClient(cc grpc.ClientConnInterface) AIServiceClient {
	return &aIServiceClient{cc}
}

func (c *aIServiceClient) Invoke(ctx context.Context, in *LLMRequest, opts ...grpc.CallOption) (*LLMResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LLMResponse)
	err := c.cc.Invoke(ctx, AIService_Invoke_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aIServiceClient) Stream(ctx context.Context, in *LLMRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[StreamEvent], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &AIService_ServiceDesc.Streams[0], AIService_Stream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[LLMRequest, StreamEvent]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AIService_StreamClient = grpc.ServerStreamingClient[StreamEvent]

// AIServiceServer is the server API for AIService service.
// All implementations must embed UnimplementedAIServiceServer
// for forward compatibility.
type AIServiceServer interface {
	Invoke(context.Context, *LLMRequest) (*LLMResponse, error)
	Stream(*LLMRequest, grpc.ServerStreamingServer[StreamEvent]) error
	mustEmbedUnimplementedAIServiceServer()
}

// UnimplementedAIServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAIServiceServer struct{}

func (UnimplementedAIServiceServer) Invoke(context.Context, *LLMRequest) (*LLMResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Invoke not implemented")
}
func (UnimplementedAIServiceServer) Stream(*LLMRequest, grpc.ServerStreamingServer[StreamEvent]) error {
	return status.Errorf(codes.Unimplemented, "method Stream not implemented")
}
func (UnimplementedAIServiceServer) mustEmbedUnimplementedAIServiceServer() {}
func (UnimplementedAIServiceServer) testEmbeddedByValue()                   {}

// UnsafeAIServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AIServiceServer will
// result in compilation errors.
type UnsafeAIServiceServer interface {
	mustEmbedUnimplementedAIServiceServer()
}

func RegisterAIServiceServer(s grpc.ServiceRegistrar, srv AIServiceServer) {
	// If the following call pancis, it indicates UnimplementedAIServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AIService_ServiceDesc, srv)
}

func _AIService_Invoke_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LLMRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServiceServer).Invoke(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AIService_Invoke_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServiceServer).Invoke(ctx, req.(*LLMRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AIService_Stream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LLMRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AIServiceServer).Stream(m, &grpc.GenericServerStream[LLMRequest, StreamEvent]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AIService_StreamServer = grpc.ServerStreamingServer[StreamEvent]

// AIService_ServiceDesc is the grpc.ServiceDesc for AIService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AIService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ai.v1.AIService",
	HandlerType: (*AIServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Invoke",
			Handler:    _AIService_Invoke_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Stream",
			Handler:       _AIService_Stream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "ai/v1/ai.proto",
}

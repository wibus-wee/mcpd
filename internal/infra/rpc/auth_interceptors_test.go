package rpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"

	"mcpv/internal/domain"
)

const bufSize = 1024 * 1024

type TestServiceServer interface {
	EmptyCall(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	StreamingOutputCall(*emptypb.Empty, TestServiceStreamingOutputCallServer) error
}

type TestServiceStreamingOutputCallServer interface {
	Send(*emptypb.Empty) error
	grpc.ServerStream
}

type testService struct{}

func (s *testService) EmptyCall(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *testService) StreamingOutputCall(_ *emptypb.Empty, stream TestServiceStreamingOutputCallServer) error {
	return stream.Send(&emptypb.Empty{})
}

type testServiceStreamingOutputCallServer struct {
	grpc.ServerStream
}

func (s *testServiceStreamingOutputCallServer) Send(msg *emptypb.Empty) error {
	return s.SendMsg(msg)
}

func registerTestServiceServer(server *grpc.Server, srv TestServiceServer) {
	server.RegisterService(&testServiceServiceDesc, srv)
}

func emptyCallClient(ctx context.Context, conn *grpc.ClientConn) error {
	out := new(emptypb.Empty)
	return conn.Invoke(ctx, "/test.TestService/EmptyCall", &emptypb.Empty{}, out)
}

func streamingOutputCallClient(ctx context.Context, conn *grpc.ClientConn) (grpc.ClientStream, error) {
	stream, err := conn.NewStream(ctx, &testServiceServiceDesc.Streams[0], "/test.TestService/StreamingOutputCall")
	if err != nil {
		return nil, err
	}
	if err := stream.SendMsg(&emptypb.Empty{}); err != nil {
		return nil, err
	}
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}
	return stream, nil
}

//nolint:revive // Signature defined by grpc handler contract.
func testServiceEmptyCallHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServiceServer).EmptyCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/test.TestService/EmptyCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServiceServer).EmptyCall(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func testServiceStreamingOutputCallHandler(srv interface{}, stream grpc.ServerStream) error {
	in := new(emptypb.Empty)
	if err := stream.RecvMsg(in); err != nil {
		return err
	}
	return srv.(TestServiceServer).StreamingOutputCall(in, &testServiceStreamingOutputCallServer{ServerStream: stream})
}

var testServiceServiceDesc = grpc.ServiceDesc{
	ServiceName: "test.TestService",
	HandlerType: (*TestServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EmptyCall",
			Handler:    testServiceEmptyCallHandler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamingOutputCall",
			Handler:       testServiceStreamingOutputCallHandler,
			ServerStreams: true,
		},
	},
	Metadata: "test.proto",
}

func TestAuthInterceptors_Token(t *testing.T) {
	authCfg := resolvedAuth{enabled: true, mode: domain.RPCAuthModeToken, token: "secret"}
	listener := bufconn.Listen(bufSize)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(authUnaryServerInterceptor(authCfg)),
		grpc.ChainStreamInterceptor(authStreamServerInterceptor(authCfg)),
	)
	registerTestServiceServer(server, &testService{})

	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	err = emptyCallClient(context.Background(), conn)
	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	ctx := metadata.AppendToOutgoingContext(context.Background(), authorizationHeader, "Bearer secret")
	err = emptyCallClient(ctx, conn)
	require.NoError(t, err)

	stream, err := streamingOutputCallClient(context.Background(), conn)
	require.NoError(t, err)
	var unauthorizedResp emptypb.Empty
	err = stream.RecvMsg(&unauthorizedResp)
	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))

	stream, err = streamingOutputCallClient(ctx, conn)
	require.NoError(t, err)
	var resp emptypb.Empty
	err = stream.RecvMsg(&resp)
	require.NoError(t, err)
}

func TestAuthClientInterceptor_Token(t *testing.T) {
	authCfg := resolvedAuth{enabled: true, mode: domain.RPCAuthModeToken, token: "secret"}
	listener := bufconn.Listen(bufSize)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(authUnaryServerInterceptor(authCfg)),
	)
	registerTestServiceServer(server, &testService{})

	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	dialer := func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(authUnaryClientInterceptor("secret")),
	)
	require.NoError(t, err)
	defer conn.Close()

	err = emptyCallClient(context.Background(), conn)
	require.NoError(t, err)
}

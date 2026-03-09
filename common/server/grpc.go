package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

// GRPCServer 封装 gRPC 服务，便于多服务复用。
type GRPCServer struct {
	Addr   string
	Server *grpc.Server
	lis    net.Listener
}

// NewGRPCServer 创建 gRPC 服务。
func NewGRPCServer(addr string, opts ...grpc.ServerOption) (*GRPCServer, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(opts...)
	return &GRPCServer{Addr: addr, Server: s, lis: lis}, nil
}

// Start 启动 gRPC 监听。
func (s *GRPCServer) Start() error {
	return s.Server.Serve(s.lis)
}

// Shutdown 优雅关闭。
func (s *GRPCServer) Shutdown(_ context.Context) {
	s.Server.GracefulStop()
}

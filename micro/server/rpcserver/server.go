package rpcserver

import (
	"context"
	"mymicro/pkg/log"
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	apimetadata "mymicro/api/metadata"
	srvintc "mymicro/micro/server/rpcserver/serverinterceptors"
	"mymicro/pkg/host"
)

type ServerOption func(o *Server)

type Server struct {
	*grpc.Server

	baseCtx            context.Context
	address            string
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcOpts           []grpc.ServerOption
	lis                net.Listener
	timeout            time.Duration

	health   *health.Server
	metadata *apimetadata.Server
	endpoint *url.URL

	enableMetrics bool
}

func (s *Server) Address() string { return s.address }

// 提取ip和端口
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	addr, err := host.Extract(s.address, s.lis)
	if err != nil {
		_ = s.lis.Close()
		return err
	}
	s.endpoint = &url.URL{
		Scheme: "grpc",
		Host:   addr,
	}
	return nil
}

func NewServer(opts ...ServerOption) *Server {
	srv := Server{
		address: ":0",
		health:  health.NewServer(),
		//timeout: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(&srv)
	}
	// 如果用户不设置拦截器，则自动默认加上一些必须的拦截器，例如：recover, timeout, tracing
	unaryInts := []grpc.UnaryServerInterceptor{
		srvintc.UnaryRecoverInterceptor,
		srvintc.UnaryTimeoutInterceptor(srv.timeout),
		otelgrpc.UnaryServerInterceptor(),
	}

	if srv.enableMetrics {
		unaryInts = append(unaryInts, srvintc.UnaryPrometheusInterceptor)
	}
	if srv.timeout > 0 {
		unaryInts = append(unaryInts, srvintc.UnaryTimeoutInterceptor(srv.timeout))
	}
	if len(srv.unaryInterceptors) > 0 {
		unaryInts = append(unaryInts, srv.unaryInterceptors...)
	}
	// 把用户传入的拦截器转换成grpc的ServerOption
	grpcOpts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(srv.unaryInterceptors...)}
	// 把用户传入的grpc.ServerOption放在一起
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}
	srv.Server = grpc.NewServer(grpcOpts...)

	// 注册metadata的Server
	srv.metadata = apimetadata.NewServer(srv.Server)

	// 解析address
	err := srv.listenAndEndpoint()
	if err != nil {
		return nil
	}

	// 注册health
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	// 支持用户通过grpc的一个接口查看当前支持的所有rpc服务
	apimetadata.RegisterMetadataServer(srv.Server, srv.metadata)
	reflection.Register(srv.Server)

	return &srv
}

func (s *Server) Start(ctx context.Context) error {
	s.baseCtx = ctx
	log.Infof("[gRPC] server listening on: %s", s.lis.Addr().String())
	s.health.Resume()
	return s.Serve(s.lis)
}

func (s *Server) Stop(_ context.Context) error {
	s.health.Shutdown()
	s.GracefulStop()
	log.Info("[gRPC] server stopped")
	return nil
}

func WithAddress(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

func WithMetrics(metric bool) ServerOption {
	return func(s *Server) {
		s.enableMetrics = metric
	}
}

func WithLis(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithUnaryInterceptor(ins ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = ins
	}
}

func WithStreamInterceptor(ins ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInterceptors = ins
	}
}

func WithOptions(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

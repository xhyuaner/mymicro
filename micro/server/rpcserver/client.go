package rpcserver

import (
	"context"
	"mymicro/micro/server/rpcserver/clientinterptors"
	"mymicro/micro/server/rpcserver/resolver/discovery"
	"mymicro/pkg/log"

	grpcInsecure "google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"
	"mymicro/micro/registry"
	"time"
)

type ClientOption func(o *clientOptions)

type clientOptions struct {
	endpoint           string
	timeout            time.Duration
	discovery          registry.Discovery
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	rpcOpts            []grpc.DialOption
	balancerName       string

	logger log.LogHelper
}

func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

func WithDiscovery(discovery registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = discovery
	}
}

func WithClientUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.unaryInterceptors = in
	}
}

func WithClientStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(o *clientOptions) {
		o.streamInterceptors = in
	}
}

func WithClientOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.rpcOpts = opts
	}
}

func WithBalancerName(name string) ClientOption {
	return func(o *clientOptions) {
		o.balancerName = name
	}
}

func DailInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dail(ctx, true, opts...)
}
func Dail(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dail(ctx, false, opts...)
}

func dail(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		timeout:      2000 * time.Millisecond,
		balancerName: "round_robin",
	}
	for _, o := range opts {
		o(&options)
	}

	// TODO 客户端默认拦截器
	ints := []grpc.UnaryClientInterceptor{
		clientinterptors.TimeoutInterceptor(options.timeout),
	}
	var streamInts []grpc.StreamClientInterceptor
	if len(options.unaryInterceptors) > 0 {
		ints = append(ints, options.unaryInterceptors...)
	}
	if len(options.streamInterceptors) > 0 {
		streamInts = append(streamInts, options.streamInterceptors...)
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "` + options.balancerName + `"}`),
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(streamInts...),
	}

	// 服务发现选项
	if options.discovery != nil {
		grpcOpts = append(grpcOpts, grpc.WithResolvers(
			discovery.NewBuilder(options.discovery, discovery.WithInsecure(insecure))),
		)
	}

	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}

	if len(options.rpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.rpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}

//func WithLogger(log *log.Logger) ClientOption {
//	return func(o *clientOptions) {
//		o.logger = log
//	}
//}

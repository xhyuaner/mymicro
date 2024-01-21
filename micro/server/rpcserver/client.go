package rpcserver

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"time"

	"google.golang.org/grpc"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"

	"mymicro/micro/registry"
	"mymicro/micro/server/rpcserver/clientinterceptors"
	"mymicro/micro/server/rpcserver/resolver/discovery"
	"mymicro/pkg/log"
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
	logger             log.LogHelper
	enableTracing      bool
}

// WithEnableTracing 是否开启链路
func WithEnableTracing(enable bool) ClientOption {
	return func(o *clientOptions) {
		o.enableTracing = enable
	}
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
		timeout:       2000 * time.Millisecond,
		balancerName:  "round_robin",
		enableTracing: true,
	}
	for _, o := range opts {
		o(&options)
	}

	// TODO 客户端默认拦截器
	ints := []grpc.UnaryClientInterceptor{
		clientinterceptors.TimeoutInterceptor(options.timeout),
	}
	if options.enableTracing {
		ints = append(ints, otelgrpc.UnaryClientInterceptor())
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

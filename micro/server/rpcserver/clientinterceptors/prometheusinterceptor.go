package clientinterceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"mymicro/micro/core/metric"
	"strconv"
	"time"
)

const serverNamespace = "rpc_client"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "xhy_duration_ms",
		Help:      "rpc server requests duration(ms).",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "xhy_code_total",
		Help:      "rpc server requests code count.",
		Labels:    []string{"method", "code"},
	})
)

func PrometheusInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		// 记录耗时
		metricServerReqDur.Observe(int64(time.Since(startTime)/time.Millisecond), method)
		// 记录状态码
		metricServerReqCodeTotal.Inc(method, strconv.Itoa(int(status.Code(err))))
		return err
	}
}

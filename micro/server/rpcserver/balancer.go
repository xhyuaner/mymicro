package rpcserver

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"

	"mymicro/micro/registry"
	"mymicro/micro/server/rpcserver/selector"
)

const (
	balancerName = "selector"
)

var (
	_ base.PickerBuilder = &balancerBuilder{}
	_ balancer.Picker    = &balancerPicker{}
)

func InitBuilder() {
	b := base.NewBalancerBuilder(
		balancerName,
		&balancerBuilder{builder: selector.GlobalSelector()},
		base.Config{HealthCheck: true},
	)
	balancer.Register(b)
}

type balancerBuilder struct {
	builder selector.Builder
}

func (b *balancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	nodes := make([]selector.Node, 0, len(info.ReadySCs))
	for conn, info := range info.ReadySCs {
		ins, _ := info.Address.Attributes.Value("rawServiceInstance").(*registry.ServiceInstance)
		nodes = append(nodes, &grpcNode{
			Node:    selector.NewNode("grpc", info.Address.Addr, ins),
			subConn: conn,
		})
	}
	p := &balancerPicker{
		selector: b.builder.Build(),
	}
	p.selector.Apply(nodes)
	return p
}

type balancerPicker struct {
	selector selector.Selector
}

func (p *balancerPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	n, done, err := p.selector.Select(info.Ctx)
	if err != nil {
		return balancer.PickResult{}, err
	}
	return balancer.PickResult{
		SubConn: n.(*grpcNode).subConn,
		Done: func(di balancer.DoneInfo) {
			done(info.Ctx, selector.DoneInfo{
				Err:           di.Err,
				BytesSent:     di.BytesSent,
				BytesReceived: di.BytesReceived,
				ReplyMD:       Trailer(di.Trailer),
			})
		},
	}, nil
}

type Trailer metadata.MD

func (t Trailer) Get(k string) string {
	v := metadata.MD(t).Get(k)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

type grpcNode struct {
	selector.Node
	subConn balancer.SubConn
}

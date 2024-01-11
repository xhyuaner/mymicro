package direct

import (
	"google.golang.org/grpc/resolver"
	"strings"
)

func init() {
	resolver.Register(NewBuilder())
}

type directBuilder struct{}

func NewBuilder() *directBuilder {
	return &directBuilder{}
}

func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addrs := make([]resolver.Address, 0)
	for _, addr := range strings.Split(strings.TrimPrefix(target.URL.Path, "/"), ",") {
		addrs = append(addrs, resolver.Address{Addr: addr})
	}
	err := cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		return nil, err
	}
	return newDirectResolver(), nil
}

func (d *directBuilder) Scheme() string {
	return "direct"
}

var _ resolver.Builder = &directBuilder{}

package discovery

import (
	"context"
	"errors"
	"google.golang.org/grpc/resolver"
	"mymicro/micro/registry"
	"strings"
	"time"
)

const name = "discovery"

type Option func(b *builder)

type builder struct {
	discoverer registry.Discovery
	timeout    time.Duration
	insecure   bool
}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	var (
		err error
		w   registry.Watcher
	)
	done := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		w, err = b.discoverer.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(b.timeout):
		err = errors.New("discovery create watcher overtime")
	}
	if err != nil {
		cancel()
		return nil, err
	}
	r := &discoveryResolver{
		w:        w,
		cc:       cc,
		cancel:   cancel,
		insecure: b.insecure,
	}
	go r.watch()
	return r, nil
}

func (b *builder) Scheme() string {
	return name
}

func NewBuilder(d registry.Discovery, opts ...Option) resolver.Builder {
	b := &builder{
		discoverer: d,
		timeout:    time.Second * 10,
		insecure:   false,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) {
		b.timeout = timeout
	}
}

func WithInsecure(insecure bool) Option {
	return func(b *builder) {
		b.insecure = insecure
	}
}

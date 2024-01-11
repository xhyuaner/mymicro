package app

import (
	"mymicro/micro/server/rpcserver"
	"net/url"
	"os"
	"time"

	"mymicro/micro/registry"
)

type Option func(o *options)

type options struct {
	id        string
	name      string
	endpoints []*url.URL
	sigs      []os.Signal
	// 允许用户传入自己的实现
	registrar        registry.Registrar
	registrarTimeout time.Duration
	stopTimeout      time.Duration

	rpcServer *rpcserver.Server
}

func WithRegistrar(registrar registry.Registrar) Option {
	return func(o *options) {
		o.registrar = registrar
	}
}

func WithEndpoints(endpoints []*url.URL) Option {
	return func(o *options) {
		o.endpoints = endpoints
	}
}

func WithRPCServer(server *rpcserver.Server) Option {
	return func(o *options) {
		o.rpcServer = server
	}
}

func WithId(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func WithSigs(sigs []os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

package app

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	"mymicro/micro/registry"
	"mymicro/pkg/log"
)

type App struct {
	opts options

	lk       sync.Mutex
	instance *registry.ServiceInstance
}

func New(opts ...Option) *App {
	// 设置默认值
	o := options{
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: 10 * time.Second,
		stopTimeout:      10 * time.Second,
	}
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}

	for _, opt := range opts {
		opt(&o)
	}
	return &App{
		opts: o,
	}
}

// Run 启动整个微服务
func (a *App) Run() error {
	// 注册的信息
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}
	// instance可能被其他协程访问到
	a.lk.Lock()
	a.instance = instance
	a.lk.Unlock()

	//if a.opts.rpcServer != nil {
	//	// 启动rpc服务
	//	a.opts.rpcServer.Server()
	//}

	//if a.opts.rpcServer != nil {
	//	err := a.opts.rpcServer.Start()
	//	if err != nil {
	//		return err
	//	}
	//}

	go func() {
		err := a.opts.rpcServer.Start(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	// 注册服务
	if a.opts.registrar != nil {
		rctx, rcancel := context.WithTimeout(context.Background(), a.opts.registrarTimeout)
		defer rcancel()
		err := a.opts.registrar.Register(rctx, instance)
		if err != nil {
			log.Errorf("register service error: %s", err)
			//fmt.Printf("register service error: %s", err)
			return err
		}
	}

	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	<-c
	return nil
}

// Stop 停止服务
func (a *App) Stop() error {
	a.lk.Lock()
	instance := a.instance
	a.lk.Unlock()

	if a.opts.registrar != nil && instance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
		defer cancel()
		if err := a.opts.registrar.Deregister(ctx, instance); err != nil {
			log.Errorf("deregister service error: %s", err)
			//fmt.Printf("deregister service error: %s", err)
			return err
		}
	}
	return nil
}

// 创建服务注册结构体
func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}

	// 从rpcServer, restServer去主动获取address信息
	if a.opts.rpcServer != nil {
		u := &url.URL{
			Scheme: "grpc",
			Host:   a.opts.rpcServer.Address(),
		}
		endpoints = append(endpoints, u.String())
	}

	// 自动生成endpoints, 自动解析ip地址，自动解析出端口
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Endpoints: endpoints,
	}, nil
}

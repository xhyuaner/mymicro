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
	"golang.org/x/sync/errgroup"

	"mymicro/micro/registry"
	ms "mymicro/micro/server"
	"mymicro/pkg/log"
)

type App struct {
	opts options

	lk       sync.Mutex
	instance *registry.ServiceInstance

	cancel func()
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

	// 启动所有服务
	var servers []ms.Server
	if a.opts.restServer != nil {
		servers = append(servers, a.opts.restServer)
	}
	if a.opts.rpcServer != nil {
		servers = append(servers, a.opts.rpcServer)
	}
	// 保证多个server之前的状态同步
	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel
	eg, ctx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}
	for _, srv := range servers {
		// 监听服务状态
		eg.Go(func() error {
			<-ctx.Done() // 等待终止信号
			// 处理stop超时
			sctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
			defer cancel()
			return srv.Stop(sctx)
		})
		wg.Add(1)
		// 启动服务
		eg.Go(func() error {
			wg.Done()
			log.Info("Start server")
			return srv.Start(ctx)
		})
	}

	wg.Wait()
	//// 启动rpcServer
	//if a.opts.rpcServer != nil {
	//	// 监听服务状态
	//	eg.Go(func() error {
	//		<-ctx.Done() // 等待终止信号
	//		// 处理stop超时
	//		sctx, cancel := context.WithTimeout(context.Background(), a.opts.stopTimeout)
	//		defer cancel()
	//		return a.opts.rpcServer.Stop(sctx)
	//	})
	//	// 启动服务
	//	eg.Go(func() error {
	//		log.Info("Start rpc server")
	//		return a.opts.rpcServer.Start(ctx)
	//	})
	//}

	// 注册服务
	if a.opts.registrar != nil {
		ctx, cancel := context.WithTimeout(context.Background(), a.opts.registrarTimeout)
		defer cancel()
		err := a.opts.registrar.Register(ctx, instance)
		if err != nil {
			log.Errorf("register service error: %s", err)
			//fmt.Printf("register service error: %s", err)
			return err
		}
	}

	// 监听退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return a.Stop()
		}
	})
	if err := eg.Wait(); err != nil {
		return err
	}
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
	if a.cancel != nil {
		a.cancel()
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

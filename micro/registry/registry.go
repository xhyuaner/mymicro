package registry

import "context"

// Registrar 服务注册接口
type Registrar interface {
	// Register 注册
	Register(ctx context.Context, service *ServiceInstance) error
	// Deregister 注销
	Deregister(ctx context.Context, service *ServiceInstance) error
}

// Discovery 服务发现接口
type Discovery interface {
	// GetService 获取服务实例
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch 创建服务监听器
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

type Watcher interface {
	// Next 获取服务实例，在以下情况会返回服务：
	// 1.第一次监听时，如果服务实例列表不为空，则返回服务实例列表
	// 2.如果服务实例发生变化，则返回服务实例列表
	// 3.如果上述都不满足，则会阻塞到context deadline或者cancel
	Next() ([]*ServiceInstance, error)
	// Stop 主动放弃监听
	Stop() error
}

type ServiceInstance struct {
	// 注册到注册中心的服务ID
	ID string `json:"id"`
	// 服务名称
	Name string `json:"name"`
	// 服务版本
	Version string `json:"version"`
	// 服务元数据
	Metadata map[string]string `json:"metadata"`

	Endpoints []string `json:"endpoints"`
}

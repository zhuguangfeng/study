package registry

import (
	"context"
	"io"
)

type Registry interface {
	Registry(ctx context.Context, si ServiceInstance) error
	UnRegistry(ctx context.Context, si ServiceInstance) error
	ListServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(ServiceName string) (<-chan Event, error)
	io.Closer
}

type ServiceInstance struct {
	Name string
	//最关键的 就是定位信息
	Address string
}

type Event struct {
	//ADD DELETE
	Type string
}

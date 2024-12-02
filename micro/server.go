package micro

import (
	"context"
	"github.com/zhuguangfeng/study/micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

type Server struct {
	name            string
	registry        registry.Registry
	registryTimeout time.Duration
	*grpc.Server
	listener net.Listener
}

func NewServer(name string, opts ...ServerOption) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registryTimeout: time.Second * 10,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func (s *Server) Start(ctx context.Context, addr string) error {

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener

	//开始注册
	if s.registry != nil {
		//在这里注册
		ctx, cancel := context.WithTimeout(context.Background(), s.registryTimeout)
		defer cancel()
		err = s.registry.Registry(ctx, registry.ServiceInstance{
			Name:    s.name,
			Address: listener.Addr().String(),
		})
		if err != nil {
			return err
		}
		//这里已经注册成功了
		defer func() {
			//忽略或者log一下错误
			//_ = s.registry.Close()
			//_ = s.registry.UnRegistry(context.Background(), registry.ServiceInstance{})
		}()
	}

	return s.Serve(listener)
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	s.GracefulStop()
	//s.listener.Close()
	return nil
}

func ServiceWithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}

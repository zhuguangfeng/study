package micro

import (
	"context"
	"github.com/zhuguangfeng/study/micro/registry"
	"google.golang.org/grpc/resolver"
	"time"
)

type grpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func (b *grpcResolverBuilder) Scheme() string {
	//TODO implement me
	panic("implement me")
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*grpcResolverBuilder, error) {
	return &grpcResolverBuilder{
		r:       r,
		timeout: timeout,
	}, nil
}

func (b *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc:      cc,
		r:       b.r,
		target:  target,
		timeout: b.timeout,
	}
	r.resolve()

	go r.watch()

	return r, nil
}

type grpcResolver struct {
	target  resolver.Target
	r       registry.Registry
	cc      resolver.ClientConn
	timeout time.Duration
	close   chan struct{}
}

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}

	for {
		select {
		case <-events:
			g.resolve()

			//case event := <-events:
			//switch event.Type {
			//case "ADD":
			//case "DELETE":
			//	//删除已有借点
			//
			//}
		case <-g.close:
			return
		}
	}

}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instances))

	for _, si := range instances {
		address = append(address, resolver.Address{Addr: si.Address})
	}

	err = g.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		g.cc.ReportError(err)
		return
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}

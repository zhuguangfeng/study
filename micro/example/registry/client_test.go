package registry

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/zhuguangfeng/study/micro"
	"github.com/zhuguangfeng/study/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)

	client, err := micro.NewClient(micro.ClientInsecure(), micro.ClientWithRegistry(r, time.Second*3))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	require.NoError(t, err)

	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	//uc:=gen.NewUserServiceClient(cc)
	//resp,err:=uc.GetById(ctx ,&gen.GetById(Id:13))
	//require.NoError(t, err)
	//t.Log(resp)

}

package cache

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/zhuguangfeng/study/cache/mocks"
	"testing"
	"time"
)

func TestClient_Lock(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		key      string
		wantErr  error
		wantLock *Lock
	}{
		{
			name: "set nx err",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key: "key1",
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, ErrFailedToPreemptLock)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key: "key1",
			},
			wantErr: ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &Lock{
				key:        "key1",
				expiration: time.Minute,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			client := NewClient(tc.mock(ctrl))

			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.Equal(t, tc.wantLock.expiration, l.expiration)
			//赋予值了
			assert.NotEmpty(t, l.val)
		})
	}
}

func TestLock_Unlock(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		key     string
		val     string
		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).Return(res)
				return cmd
			},
			key:     "key1",
			val:     "val1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).Return(res)
				return cmd
			},
			key:     "key1",
			val:     "val1",
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), luaUnlock, []string{"key1"}, []any{"val1"}).Return(res)
				return cmd
			},
			key:     "key1",
			val:     "val1",
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			lock := &Lock{
				client: tc.mock(ctrl),
				key:    tc.key,
				val:    tc.val,
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

// 刷新锁的示例
func ExampleLock_Refresh() {
	var l *Lock
	stopChan := make(chan struct{})
	errChan := make(chan error)
	timeoutChan := make(chan struct{}, 1)
	//续约
	go func() {
		ticker := time.NewTicker(time.Second * 10)

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := l.Refresh(ctx)
				cancel()
				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					close(stopChan)
					close(errChan)
					return
				}
			case <-stopChan:
				//l.Unlock(context.Background())
				return
			}
		}

	}()

	//假设这是是你的业务 循环处理的逻辑
	for i := 0; i < 100; i++ {
		select {
		case <-errChan:
			break
		default:
			//正常业务逻辑
		}
	}

	//如果你的业务不是循环处理 那就要每个步骤检测一下
	select {
	case <-errChan:
	//续约失败 要中断业务
	default:
		//正常执行业务逻辑
	}

	//业务结束要退出续约的循环
	stopChan <- struct{}{}
	l.Unlock(context.Background())

	//下面内容删掉就不可以运行这个方法了
	fmt.Println("Hello")
	// Output:
	// Hello

}

func ExampleLock_AutoRefresh() {
	var l *Lock

	go func() {
		//这里返回error了 你要中断业务
		l.AutoRefresh(time.Second*30, time.Second)
	}()

	fmt.Println("Hello")
	// Output:
	// Hello
}

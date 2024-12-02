package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold         = errors.New("redis-lock: 锁不存在")

	//go:embed lua/unlock.lua
	luaUnlock string
	//go:embed lua/refresh.lua
	luaRefresh string
	//go:embed lua/lock.lua
	luaLock string
)

// Client就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration, timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	var timer *time.Timer
	val := uuid.New().String()
	for {

		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if res == "OK" {
			return &Lock{
				client:     c.client,
				key:        key,
				val:        val,
				expiration: expiration,
				unlockChan: make(chan struct{}, 1),
			}, nil
		}

		//在这里重试
		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	val := uuid.New().String()

	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		// 别人抢到了锁
		return nil, ErrFailedToPreemptLock
	}
	return &Lock{
		client:     c.client,
		key:        key,
		val:        val,
		expiration: expiration,
		unlockChan: make(chan struct{}, 1),
	}, nil
}

//func (c *Client) Unlock(ctx context.Context, lock *Lock) error {
//
//}

type Lock struct {
	client     redis.Cmdable
	key        string
	val        string
	expiration time.Duration
	unlockChan chan struct{}
}

// 自动续约
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)
	//续约

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				timeoutChan <- struct{}{}
				continue
			}
			if err != nil {
				return err
			}

		case <-l.unlockChan:
			return nil
		}
	}

}

// 手动续约
func (l *Lock) Refresh(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.val, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	return nil
}

// 解锁
func (l *Lock) Unlock(ctx context.Context) error {
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.val).Int64()
	defer func() {
		select {
		case l.unlockChan <- struct{}{}:
		default:
			//说明没有人调用 AutoRefresh
		}

	}()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	return nil
}

// 错误写法 因为 get和del中间有间隔 可能会被其它实例插入
//func (l *Lock) Unlock(ctx context.Context) error {
//	//先要判断一下这把锁是不是自己的锁
//	val, err := l.client.Get(ctx, l.key).Result()
//	if err != nil {
//		return err
//	}
//	if val != l.val {
//		return errors.New("锁不是自己的")
//	}
//	cnt, err := l.client.Del(ctx, l.key).Result()
//	if err != nil {
//		return err
//	}
//	if cnt != 1 {
//		//这个地方代表你加的锁过期了
//		return ErrLockNotExist
//	}
//
//	return nil
//}

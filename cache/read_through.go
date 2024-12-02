package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	errFailedToRefreshCache = errors.New("刷新缓存失败")
)

type ReadThroughCache struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration
}

// 同步
func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err := r.LoadFunc(ctx, key)
		if err == nil {
			er := r.Cache.Set(ctx, key, val, r.Expiration)
			if er != nil {
				return val, fmt.Errorf("%w, 原因: %s", errFailedToRefreshCache, er.Error())
			}
		}
	}
	return val, err
}

// 全异步
func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		go func() {
			val, err := r.LoadFunc(ctx, key)
			if err == nil {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}
		}()
	}
	return val, err
}

// 办异步
func (r *ReadThroughCache) GetV2(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err := r.LoadFunc(ctx, key)
		if err == nil {
			go func() {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}()
		}
	}
	return val, err
}

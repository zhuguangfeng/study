package cache

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	errKeyNotFound = errors.New("cache: 键不存在")
	errKeyExpired  = errors.New("键过期")
)

type item struct {
	val      any
	deadline time.Time
}

type BuildInMapCacheOption func(cache *BuildInMapCache)

// 本地缓存结构
type BuildInMapCache struct {
	data      map[string]*item
	mutex     sync.RWMutex
	close     chan struct{}
	onEvicted func(key string, value any)
	//onEvicted func(ctx context.Context, key string, val any)
	//onEvicteds []func(key string, val any)
}

func NewBuildInMapCache(interval time.Duration, opts ...BuildInMapCacheOption) *BuildInMapCache {
	res := &BuildInMapCache{
		data:  make(map[string]*item),
		close: make(chan struct{}),
		mutex: sync.RWMutex{},
		onEvicted: func(key string, value any) {

		},
	}

	for _, opt := range opts {
		opt(res)
	}

	// 定时清理一定数量的过期key
	go func() {
		ticker := time.NewTicker(interval)

		for {
			select {
			case t := <-ticker.C:
				res.mutex.Lock()
				i := 0
				for key, val := range res.data {
					if i > 1000 {
						break
					}
					if val.deadlineBefore(t) {
						res.delete(key)
					}
					i++
				}
			case <-res.close:
				return
			}
		}
	}()

	return res
}

// 设置缓存
func (c *BuildInMapCache) Set(key string, val any, expiration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.set(key, val, expiration)

}
func (c *BuildInMapCache) set(key string, val any, expiration time.Duration) error {
	var dl time.Time
	if expiration > 0 {
		dl = time.Now().Add(expiration)
	}
	c.data[key] = &item{
		val:      val,
		deadline: dl,
	}
	return nil
}

// 获取缓存
func (c *BuildInMapCache) Get(key string) (any, error) {
	c.mutex.Lock()
	res, ok := c.data[key]
	c.mutex.Unlock()
	if !ok {
		return nil, fmt.Errorf("%w,key:%s", errKeyNotFound, key)
	}
	now := time.Now()
	if res.deadlineBefore(now) {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		res, ok = c.data[key]
		if res.deadlineBefore(now) {
			c.delete(key)
			return nil, fmt.Errorf("%w,key:%s", errKeyNotFound, key)
		}
	}
	return res.val, nil
}

// 删除缓存
func (c *BuildInMapCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.delete(key)
	return nil
}

func (c *BuildInMapCache) delete(key string) {
	item, ok := c.data[key]
	if !ok {
		return
	}
	delete(c.data, key)
	c.onEvicted(key, item.val)
}

func (c *BuildInMapCache) Close() error {
	c.close <- struct{}{}
	return nil
}

// 判断是否过期
func (i *item) deadlineBefore(t time.Time) bool {
	return !i.deadline.IsZero() && i.deadline.Before(t)
}

func BuildInMapCacheWithEvictedCallback(fn func(key string, val any)) BuildInMapCacheOption {
	return func(cache *BuildInMapCache) {
		cache.onEvicted = fn
	}
}

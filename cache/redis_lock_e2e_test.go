//go:build e2e

package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestClient_e2e_TryLock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		key        string
		expiration time.Duration
		wantErr    error
		wantLock   *Lock
	}{
		{
			//别人持有锁了
			name: "key exist",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "key1", "val1", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key1").Result()
				require.NoError(t, err)
				assert.Equal(t, "val1", res)
			},

			key:        "key1",
			expiration: time.Minute,
			wantErr:    ErrFailedToPreemptLock,
		},
		{
			//加锁成功
			name:   "locked",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "key2").Result()
				require.NoError(t, err)
				assert.NotEmpty(t, res)
			},

			key:        "key2",
			expiration: time.Minute,
			wantLock: &Lock{
				key: "key2",
				//client: rdb,
			},
		},
	}

	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			lock, err := client.TryLock(ctx, tc.key, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.NotEmpty(t, lock.val)
			//assert.NotNil(t, lock.client)
			tc.after(t)
		})
	}
}

func TestLock_e2e_Unlock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		lock    *Lock
		wantErr error
	}{
		{
			name: "lock not hold",
			before: func(t *testing.T) {

			},
			after: func(f *testing.T) {

			},
			lock: &Lock{
				key:    "unlock_key1",
				val:    "val1",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "lock hold by others",
			before: func(t *testing.T) {
				// 模拟别人的锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key2", "val2", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(f *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.GetDel(ctx, "unlock_key2").Result()
				require.NoError(t, err)
				assert.Equal(t, "val2", res)
			},
			lock: &Lock{
				key:    "unlock_key2",
				val:    "val",
				client: rdb,
			},
			wantErr: ErrLockNotHold,
		},
		{
			name: "unlocked",
			before: func(t *testing.T) {
				// 模拟自己加的锁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "unlock_key3", "val3", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(f *testing.T) {
				//锁被释放
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Exists(ctx, "unlock_key3").Result()
				require.NoError(t, err)
				assert.Equal(t, int64(0), res)
			},
			lock: &Lock{
				key:    "unlock_key3",
				val:    "val3",
				client: rdb,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			err := tc.lock.Unlock(ctx)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestClient_e2e_Lock(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		key        string
		expiration time.Duration
		timeout    time.Duration
		retry      RetryStrategy

		wantLock *Lock
		wantErr  error
	}{
		{
			name: "locked",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key1").Result()
				require.NoError(t, err)
				require.True(t, timeout >= time.Second*50)
				_, err = rdb.Del(ctx, "lock_key1").Result()
				require.NoError(t, err)
			},

			key:        "lock_key1",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &fixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key:        "lock_key1",
				expiration: time.Minute,
			},
		},

		{
			name: "others hold lock",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "lock_key2", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				res, err := rdb.GetDel(ctx, "lock_key2").Result()
				require.NoError(t, err)
				require.Equal(t, "123", res)
			},

			key:        "lock_key2",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &fixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   3,
			},
			wantErr: fmt.Errorf("redis-lock: 超出重试限制, %w", ErrFailedToPreemptLock),
		},

		{
			name: "retry and locked",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				res, err := rdb.Set(ctx, "lock_key3", "123", time.Minute).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", res)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				timeout, err := rdb.TTL(ctx, "lock_key3").Result()
				require.NoError(t, err)
				require.True(t, timeout >= time.Second*50)
				_, err = rdb.Del(ctx, "lock_key3").Result()
				require.NoError(t, err)
			},

			key:        "lock_key3",
			expiration: time.Minute,
			timeout:    time.Second * 3,
			retry: &fixedIntervalRetryStrategy{
				Interval: time.Second,
				MaxCnt:   10,
			},
			wantLock: &Lock{
				key:        "lock_key3",
				expiration: time.Minute,
			},
		},
	}

	client := NewClient(rdb)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			lock, err := client.Lock(context.Background(), tc.key, tc.expiration, tc.timeout, tc.retry)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.key, lock.key)
			assert.Equal(t, tc.wantLock.expiration, lock.expiration)
			assert.NotEmpty(t, lock.val)
			assert.NotNil(t, lock.client)
			tc.after(t)
		})
	}
}

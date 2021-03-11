/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package zcache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/zlyuancn/zcache/cachedb/memory-cache"
	"github.com/zlyuancn/zcache/loader"
	single_sf "github.com/zlyuancn/zcache/single_flight/single-sf"
	"github.com/zlyuancn/zcache/wrap_call"

	"github.com/zlyuancn/zcache/codec"
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	"github.com/zlyuancn/zcache/logger"
)

const (
	// 默认是否在缓存错误时直接返回
	defaultDirectReturnOnCacheFault = true
	// 默认是否在注册时检测到加载器存在时panic
	defaultPanicOnLoaderExists = true
)

type Cache struct {
	cache                    core.ICacheDB // 缓存数据库
	defaultExpire, maxExpire time.Duration // 默认过期时间
	directReturnOnCacheFault bool          // 在缓存故障时直接返回

	codec core.ICodec // 编解码器

	loaders             map[string]core.ILoader // 加载器注册表
	panicOnLoaderExists bool                    // 注册加载器时如果加载器已存在会panic, 设为false会替换旧的加载器
	loaderLock          sync.RWMutex            // 加载器的锁
	sf                  core.ISingleFlight      // 单跑模块

	log core.ILogger // 日志
}

func NewCache(opts ...Option) *Cache {
	c := &Cache{
		directReturnOnCacheFault: defaultDirectReturnOnCacheFault,

		codec: codec.DefaultCodec,

		loaders:             make(map[string]core.ILoader),
		panicOnLoaderExists: defaultPanicOnLoaderExists,
	}

	for _, o := range opts {
		o(c)
	}

	if c.cache == nil {
		c.cache = memory_cache.NewMemoryCache()
	}
	if c.sf == nil {
		c.sf = single_sf.NewSingleFlight()
	}
	if c.log == nil {
		c.log = logger.NoLog()
	}
	return c
}

// 注册加载器
func (c *Cache) RegisterLoader(bucket string, loader core.ILoader) {
	if bucket == "" {
		panic(errors.New("bucket name is empty"))
	}

	c.loaderLock.Lock()
	if c.loaders[bucket] != nil && c.panicOnLoaderExists {
		c.loaderLock.Unlock()
		panic("loader is exists")
	}
	c.loaders[bucket] = loader
	c.loaderLock.Unlock()
}

// 注册加载函数, 效果等同于注册加载器
func (c *Cache) RegisterLoaderFn(bucket string, fn loader.LoaderFn, opts ...loader.Option) {
	l := loader.NewLoader(fn, opts...)
	c.RegisterLoader(bucket, l)
}

// 设置数据到缓存
func (c *Cache) Set(query core.IQuery, a interface{}, ex ...time.Duration) error {
	return c.SetWithContext(nil, query, a, ex...)
}

// 设置数据到缓存
//
// ex < 0 表示永不过期, ex = 0 或未设置表示使用默认过期时间
func (c *Cache) SetWithContext(ctx context.Context, query core.IQuery, a interface{}, ex ...time.Duration) error {
	return c.doWithContext(ctx, func() error {
		return c.set(query, a, ex...)
	})
}

func (c *Cache) set(query core.IQuery, a interface{}, ex ...time.Duration) error {
	bs, err := c.marshal(a)
	if err != nil {
		query.SetError(err)
		return err
	}

	err = c.cache.Set(query, bs, c.makeExpire(ex...))
	if err != nil {
		err = fmt.Errorf("write to cache error: %s", err)
		query.SetError(err)
		return err
	}
	return nil
}

// 保存一条数据到缓存
func (c *Cache) Save(bucket string, a interface{}, ex time.Duration, queryConfig ...*QueryConfig) error {
	return c.SaveWithContext(nil, bucket, a, ex, queryConfig...)
}

// 保存一条数据到缓存
func (c *Cache) SaveWithContext(ctx context.Context, bucket string, a interface{}, ex time.Duration, queryConfig ...*QueryConfig) error {
	var query core.IQuery
	if len(queryConfig) > 0 {
		query = queryConfig[0].Bucket(bucket).Make()
	} else {
		query = NewQuery(bucket)
	}
	return c.doWithContext(ctx, func() error {
		return c.set(query, a, ex)
	})
}

// 删除指定数据
func (c *Cache) Remove(queries ...core.IQuery) error {
	return c.RemoveWithContext(nil, queries...)
}

// 删除指定数据
func (c *Cache) RemoveWithContext(ctx context.Context, queries ...core.IQuery) (err error) {
	if len(queries) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		err := c.cache.Del(queries...)
		if err == nil {
			return nil
		}
		for _, q := range queries {
			q.SetError(err)
		}
		return err
	})
}

// 删除指定数据
func (c *Cache) Del(bucket string, queryConfigs ...*QueryConfig) error {
	return c.DelWithContext(nil, bucket, queryConfigs...)
}

// 删除指定数据
func (c *Cache) DelWithContext(ctx context.Context, bucket string, queryConfigs ...*QueryConfig) error {
	if len(queryConfigs) == 0 {
		return nil
	}
	queries := make([]core.IQuery, len(queryConfigs))
	for i, config := range queryConfigs {
		queries[i] = config.Bucket(bucket).Make()
	}
	return c.doWithContext(ctx, func() error {
		err := c.cache.Del(queries...)
		if err == nil {
			return nil
		}

		for _, qc := range queryConfigs {
			qc.setError(err)
		}
		return err
	})
}

// 删除命名空间下所有数据
func (c *Cache) DelBucket(buckets ...string) error {
	return c.DelBucketWithContext(nil, buckets...)
}

// 删除命名空间下所有数据
func (c *Cache) DelBucketWithContext(ctx context.Context, buckets ...string) error {
	if len(buckets) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		return c.cache.DelBucket(buckets...)
	})
}

// 将数据解码到a
func (c *Cache) marshal(a interface{}) ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	bs, err := c.codec.Encode(a)
	if err != nil {
		return nil, fmt.Errorf("<%T> is can't encode: %s", a, err)
	}
	return bs, nil
}

// 将数据解码到a
func (c *Cache) unmarshal(bs []byte, a interface{}) error {
	if len(bs) == 0 && c.codec != codec.Byte {
		return errs.DataIsNil
	}
	err := c.codec.Decode(bs, a)
	if err != nil {
		return fmt.Errorf("can't decode to <%T>: %s", a, err)
	}
	return nil
}

// 为一个执行添加上下文
func (c *Cache) doWithContext(ctx context.Context, fn func() error) (err error) {
	if ctx == nil || ctx == context.Background() || ctx == context.TODO() {
		return wrap_call.WrapCall(fn)
	}

	done := make(chan struct{}, 1)
	go func() {
		err = wrap_call.WrapCall(fn)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// 构建超时
//
// 如果没有传入有效的expire, 使用默认的过期时间.
// 默认的过期时间会检查如果 maxExpire 有效使用 defaultExpire 和 maxExpire 区间的随机值
func (c *Cache) makeExpire(ex ...time.Duration) time.Duration {
	if len(ex) > 0 && ex[0] != 0 {
		return ex[0]
	}
	if c.maxExpire > 0 && c.defaultExpire > 0 {
		return time.Duration(rand.Int63())%(c.maxExpire-c.defaultExpire) + (c.defaultExpire)
	}
	return c.defaultExpire
}

// 关闭
func (c *Cache) Close() error {
	return c.cache.Close()
}

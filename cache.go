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
	"hash/fnv"
	"math/rand"
	"sync"
	"time"

	"github.com/zlyuancn/zcache/cachedb/memory-cache"
	"github.com/zlyuancn/zcache/loader"
	"github.com/zlyuancn/zcache/query"
	single_sf "github.com/zlyuancn/zcache/single_flight/single-sf"
	"github.com/zlyuancn/zcache/wrap_call"

	"github.com/zlyuancn/zcache/codec"
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	"github.com/zlyuancn/zcache/logger"
)

const (
	defaultDirectReturnOnCacheFault = true
	defaultPanicOnLoaderExists      = true
)

type Cache struct {
	cache                    core.ICacheDB // 缓存数据库
	startEx, endEx           time.Duration // 默认过期时间
	directReturnOnCacheFault bool          // 在缓存故障时直接返回

	codec core.ICodec // 编解码器

	loaders             map[uint64]core.ILoader // 加载器配置
	panicOnLoaderExists bool                    // 注册加载器时如果加载器已存在会panic, 设为false会替换旧的加载器
	mx                  sync.RWMutex            // 对注册的加载器加锁
	sf                  core.ISingleFlight      // 单跑模块

	log core.ILogger // 日志
}

func NewCache(opts ...Option) *Cache {
	c := &Cache{
		startEx:                  0,
		endEx:                    0,
		directReturnOnCacheFault: defaultDirectReturnOnCacheFault,

		codec: codec.DefaultCodec,

		loaders:             make(map[uint64]core.ILoader),
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
func (c *Cache) RegisterLoader(namespace, key string, loader core.ILoader) {
	if namespace == "" {
		panic(errors.New("namespace is empty"))
	}
	if key == "" {
		panic(errors.New("key is empty"))
	}

	loaderId := c.makeLoaderId(namespace, key)
	c.mx.Lock()
	if c.loaders[loaderId] != nil && c.panicOnLoaderExists {
		c.mx.Unlock()
		panic("loader is exists")
	}
	c.loaders[loaderId] = loader
	c.mx.Unlock()
}

// 注册加载函数, 效果等同于注册加载器
func (c *Cache) RegisterLoaderFn(namespace, key string, fn loader.LoaderFn, opts ...loader.Option) {
	l := loader.NewLoader(fn, opts...)
	c.RegisterLoader(namespace, key, l)
}

// 获取加载器
func (c *Cache) getLoader(namespace, key string) core.ILoader {
	loaderId := c.makeLoaderId(namespace, key)
	c.mx.RLock()
	loader := c.loaders[loaderId]
	c.mx.RUnlock()
	return loader
}

// 获取数据
func (c *Cache) Get(query core.IQuery, a interface{}) error {
	return c.GetWithContext(nil, query, a)
}

// 获取数据
func (c *Cache) GetWithContext(ctx context.Context, query core.IQuery, a interface{}) error {
	return c.doWithContext(ctx, func() error {
		return c.get(query, a)
	})
}
func (c *Cache) get(query core.IQuery, a interface{}) error {
	// 从缓存获取数据
	bs, cacheErr := c.cache.Get(query)
	if cacheErr == nil {
		return c.unmarshal(bs, a)
	}
	if cacheErr != errs.CacheMiss { // 非缓存未命中错误
		if c.directReturnOnCacheFault { // 直接报告错误
			cacheErr = fmt.Errorf("load from cache error. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.ArgsText(), cacheErr)
			return cacheErr
		}
		cacheErr = fmt.Errorf("load from cache error, The data will be fetched from the loader. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.ArgsText(), cacheErr)
		c.log.Error(cacheErr)
	}

	// 从加载器获取数据
	bs, err := c.sf.Do(query, c.load)
	if err != nil {
		return err
	}

	return c.unmarshal(bs, a)
}

// 获取数据
func (c *Cache) Query(namespace, key string, a interface{}, opts ...query.Option) error {
	return c.QueryWithContext(nil, namespace, key, a, opts...)
}

// 获取数据
func (c *Cache) QueryWithContext(ctx context.Context, namespace, key string, a interface{}, opts ...query.Option) error {
	q := NewQuery(namespace, key, opts...)
	return c.doWithContext(ctx, func() error {
		return c.get(q, a)
	})
}

// 加载数据并写入缓存
func (c *Cache) load(query core.IQuery) (bs []byte, err error) {
	err = wrap_call.WrapCall(func() error {
		// 获取加载器
		l := query.Loader() // 查询加载器的优先级高于注册表的加载器
		if l == nil {
			l = c.getLoader(query.Namespace(), query.Key()) // 没有查询加载器时从注册表中获取加载器
		}
		if l == nil {
			return errs.LoaderNotFound
		}
		// 加载数据
		result, err := l.Load(query)
		if err != nil {
			return fmt.Errorf("load data error from loader. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.ArgsText(), err)
		}

		// 编码
		bs, err = c.marshal(result)
		if err != nil {
			return err
		}

		// 写入缓存
		cacheErr := c.cache.Set(query, bs, c.makeExpire(l.Expire()))
		if cacheErr != nil {
			cacheErr = fmt.Errorf("write to cache error. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.ArgsText(), cacheErr)
			if c.directReturnOnCacheFault {
				return cacheErr
			}
			c.log.Error(cacheErr)
		}
		return nil
	})
	return bs, err
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
		bs, err := c.marshal(a)
		if err != nil {
			return err
		}

		err = c.cache.Set(query, bs, c.makeExpire(ex...))
		if err != nil {
			return fmt.Errorf("write to cache error, query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.ArgsText(), err)
		}
		return nil
	})
}

// 删除指定数据
func (c *Cache) Del(queries ...core.IQuery) error {
	return c.DelWithContext(nil, queries...)
}

// 删除指定数据
func (c *Cache) DelWithContext(ctx context.Context, queries ...core.IQuery) (err error) {
	if len(queries) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		return c.cache.Del(queries...)
	})
}

// 删除命名空间下所有数据
func (c *Cache) DelNamespace(namespaces ...string) error {
	return c.DelSpaceWithContext(nil, namespaces...)
}

// 删除命名空间下所有数据
func (c *Cache) DelSpaceWithContext(ctx context.Context, namespaces ...string) error {
	if len(namespaces) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		return c.cache.DelNamespace(namespaces...)
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

// 构建加载器id
func (c *Cache) makeLoaderId(namespace, key string) uint64 {
	f := fnv.New64a()
	_, _ = f.Write([]byte(namespace))
	_, _ = f.Write([]byte{':'})
	_, _ = f.Write([]byte(key))
	return f.Sum64()
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
func (c *Cache) makeExpire(ex ...time.Duration) time.Duration {
	if len(ex) > 0 && ex[0] != 0 {
		return ex[0]
	}
	if c.endEx > 0 && c.startEx > 0 {
		return time.Duration(rand.Int63())%(c.endEx-c.startEx) + (c.startEx)
	}
	return c.startEx
}

// 关闭
func (c *Cache) Close() error {
	return c.cache.Close()
}

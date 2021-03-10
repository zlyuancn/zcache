/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/10
   Description :
-------------------------------------------------
*/

package zcache

import (
	"context"
	"fmt"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	"github.com/zlyuancn/zcache/query"
	"github.com/zlyuancn/zcache/wrap_call"
)

// 获取加载器, 加载器不存在时返回nil
func (c *Cache) getLoader(bucket string) core.ILoader {
	c.loaderLock.RLock()
	l := c.loaders[bucket]
	c.loaderLock.RUnlock()
	return l
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
			cacheErr = fmt.Errorf("load from cache error. query: %s, args: %s, err: %s", query.Bucket(), query.ArgsText(), cacheErr)
			return cacheErr
		}
		cacheErr = fmt.Errorf("load from cache error, The data will be fetched from the loader. query: %s, args: %s, err: %s", query.Bucket(), query.ArgsText(), cacheErr)
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
//
// QueryConfig 的 bucket 会被替换传入的 bucket
func (c *Cache) Query(bucket string, a interface{}, queryConfig ...*QueryConfig) error {
	return c.QueryWithContext(nil, bucket, a, queryConfig...)
}

// 获取数据
func (c *Cache) QueryWithContext(ctx context.Context, bucket string, a interface{}, queryConfig ...*QueryConfig) error {
	var q core.IQuery
	if len(queryConfig) > 0 {
		q = queryConfig[0].Bucket(bucket).Make()
	} else {
		q = query.NewQuery(bucket)
	}
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
			l = c.getLoader(query.Bucket()) // 没有查询加载器时从注册表中获取加载器
		}
		if l == nil {
			return errs.LoaderNotFound
		}

		// 加载数据
		result, err := l.Load(query)
		if err != nil {
			return fmt.Errorf("load data error from loader. query: %s, args: %s, err: %s", query.Bucket(), query.ArgsText(), err)
		}

		// 编码
		bs, err = c.marshal(result)
		if err != nil {
			return err
		}

		// 写入缓存
		cacheErr := c.cache.Set(query, bs, c.makeExpire(l.Expire()))
		if cacheErr != nil {
			cacheErr = fmt.Errorf("write to cache error. query: %s, args: %s, err: %s", query.Bucket(), query.ArgsText(), cacheErr)
			if c.directReturnOnCacheFault {
				return cacheErr
			}
			c.log.Error(cacheErr)
		}
		return nil
	})
	return bs, err
}

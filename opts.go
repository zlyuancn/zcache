/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package zcache

import (
	"time"

	"github.com/zlyuancn/zcache/cachedb/no-cache"
	"github.com/zlyuancn/zcache/codec"
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/logger"
	no_sf "github.com/zlyuancn/zcache/single_flight/no-sf"
)

type Option func(c *Cache)

// 设置缓存数据库
func WithCacheDB(cacheDB core.ICacheDB) Option {
	return func(c *Cache) {
		if cacheDB == nil {
			cacheDB = no_cache.NoCache()
		}
		c.cache = cacheDB
	}
}

// 设置全局默认过期时间
//
// 如果 endEx > 0 且 ex > 0, 则过期时间在 [ex, endEx-1] 区间随机
// 如果 ex <= 0 (默认), 则永不过期
func WithDefaultExpire(ex time.Duration, endEx ...time.Duration) Option {
	return func(l *Cache) {
		l.startEx, l.endEx = ex, 0
		if len(endEx) > 0 {
			l.endEx = endEx[0]
		}
	}
}

// 在缓存故障时直接返回缓存错误(默认)
func WithDirectReturnOnCacheFault(b ...bool) Option {
	return func(c *Cache) {
		c.directReturnOnCacheFault = len(b) == 0 || b[0]
	}
}

// 注册加载器时如果加载器已存在会panic(默认), 设为false会替换旧的加载器
func WithPanicOnLoaderExists(err ...bool) Option {
	return func(m *Cache) {
		m.panicOnLoaderExists = len(err) == 0 || err[0]
	}
}

// 设置编码器
func WithCodec(c core.ICodec) Option {
	return func(cache *Cache) {
		if c == nil {
			c = codec.DefaultCodec
		}
		cache.codec = c
	}
}

// 设置单跑模块
func WithSingleFlight(sf core.ISingleFlight) Option {
	return func(c *Cache) {
		if sf == nil {
			sf = no_sf.NoSingleFlight()
		}
		c.sf = sf
	}
}

// 设置日志组件
func WithLogger(log core.ILogger) Option {
	return func(m *Cache) {
		if log == nil {
			log = logger.NoLog()
		}
		m.log = log
	}
}

/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package redis_cache

import (
	"time"
)

type Option func(r *redisCache)

// 设置key前缀
func WithKeyPrefix(prefix string) Option {
	return func(r *redisCache) {
		r.keyPrefix = prefix
	}
}

// 设置参数分隔符
func WithArgsSep(sep string) Option {
	return func(r *redisCache) {
		r.argsSep = sep
	}
}

// 设置操作超时时间
func WithDoTimeout(timeout time.Duration) Option {
	return func(r *redisCache) {
		r.doTimeout = timeout
	}
}

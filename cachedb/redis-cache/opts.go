/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package redis_cache

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

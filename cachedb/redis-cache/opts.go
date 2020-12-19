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

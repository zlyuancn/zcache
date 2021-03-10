/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package core

import (
	"time"
)

// 缓存数据库接口
type ICacheDB interface {
	// 设置一个值, expire <= 0 时表示永不过期
	Set(query IQuery, bs []byte, expire time.Duration) error
	// 获取一个值, 如果缓存未命中请返回 errs.CacheMiss 错误
	Get(query IQuery) ([]byte, error)
	// 获取多个值, 返回数据和错误的数量必须和请求数量一致
	MGet(queries ...IQuery) ([][]byte, []error)

	// 删除数据
	Del(queries ...IQuery) error
	// 删除bucket
	DelBucket(buckets ...string) error

	// 关闭
	Close() error
}

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
	// 设置一个值, ex <= 0 时表示永不过期
	Set(query IQuery, bs []byte, ex time.Duration) error
	// 获取一个值
	Get(query IQuery) ([]byte, error)
	// 获取多个值, 返回数量一定和请求数量一致
	MGet(queries ...IQuery) ([][]byte, []error)

	// 删除key
	Del(queries ...IQuery) error
	// 删除命名空间
	DelNamespace(namespaces ...string) error

	// 关闭
	Close() error
}

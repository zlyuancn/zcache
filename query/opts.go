/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package query

import (
	"github.com/zlyuancn/zcache/core"
)

type Option func(q *Query)

// 设置参数
func WithArgs(args interface{}) Option {
	return func(q *Query) {
		q.args = args
	}
}

// 设置元数据
func WithMeta(meta interface{}) Option {
	return func(q *Query) {
		q.meta = meta
	}
}

// 设置查询加载器, 无数据时优先使用这个加载器
func WithLoader(loader core.ILoader) Option {
	return func(q *Query) {
		q.loader = loader
	}
}

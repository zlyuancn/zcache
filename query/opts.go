/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package query

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

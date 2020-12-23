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

// 主动设置参数解析后的文本
func WithArgsText(text string) Option {
	return func(q *Query) {
		q.argsText = &text
	}
}

// 设置元数据
func WithMeta(meta interface{}) Option {
	return func(q *Query) {
		q.meta = meta
	}
}

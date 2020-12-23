/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package query

import (
	"errors"
	"hash/fnv"

	"github.com/zlyuancn/zcache/core"
)

var _ core.IQuery = (*Query)(nil)

type Query struct {
	// 空间名
	namespace string
	// 查询key
	key string
	// 查询参数
	args interface{}

	// 全局唯一id
	globalId uint64
	// 请求路径, 由 key 和 args 生成
	argsText *string

	// 元数据
	meta interface{}
}

// 创建一个查询
func NewQuery(namespace, key string, opts ...Option) core.IQuery {
	if namespace == "" {
		panic(errors.New("namespace is empty"))
	}
	if key == "" {
		panic(errors.New("key is empty"))
	}

	q := &Query{
		namespace: namespace,
		key:       key,
	}
	for _, o := range opts {
		o(q)
	}

	var bs []byte
	if q.argsText == nil {
		bs, _ = Marshal(q.args)
		var s = string(bs)
		q.argsText = &s
	} else {
		bs = []byte(*q.argsText)
	}

	f := fnv.New64a()
	_, _ = f.Write([]byte(q.namespace))
	_, _ = f.Write([]byte{':'})
	_, _ = f.Write([]byte(q.key))
	if len(bs) > 0 {
		_, _ = f.Write([]byte{'?'})
		_, _ = f.Write(bs)
	}
	q.globalId = f.Sum64()
	return q
}

func (q *Query) Namespace() string {
	return q.namespace
}
func (q *Query) Key() string {
	return q.key
}
func (q *Query) Args() interface{} {
	return q.args
}
func (q *Query) Meta() interface{} {
	return q.meta
}

func (q *Query) GlobalId() uint64 {
	return q.globalId
}
func (q *Query) ArgsText() string {
	return *q.argsText
}

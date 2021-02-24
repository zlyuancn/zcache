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
	// 参数
	args interface{}
	// 查询参数文本, 根据args生成
	argsText string

	// 全局唯一id
	globalId uint64

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

	q.makeArgsText()
	return q
}

func (q *Query) makeArgsText() {
	bs, _ := Marshal(q.args)
	q.argsText = string(bs)

	f := fnv.New64a()
	_, _ = f.Write([]byte(q.namespace))
	_, _ = f.Write([]byte{':'})
	_, _ = f.Write([]byte(q.key))
	if len(bs) > 0 {
		_, _ = f.Write([]byte{'?'})
		_, _ = f.Write(bs)
	}
	q.globalId = f.Sum64()
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
func (q *Query) ArgsText() string {
	return q.argsText
}
func (q *Query) Meta() interface{} {
	return q.meta
}

func (q *Query) GlobalId() uint64 {
	return q.globalId
}

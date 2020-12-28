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
	args string

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

	if q.globalId == 0 {
		q.makeArgs(nil)
	}
	return q
}

func (q *Query) makeArgs(args interface{}) {
	bs, _ := Marshal(args)
	q.args = string(bs)

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
func (q *Query) Args() string {
	return q.args
}
func (q *Query) Meta() interface{} {
	return q.meta
}

func (q *Query) GlobalId() uint64 {
	return q.globalId
}

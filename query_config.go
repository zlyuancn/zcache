/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/24
   Description :
-------------------------------------------------
*/

package zcache

import (
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/loader"
	"github.com/zlyuancn/zcache/query"
)

type QueryConfig struct {
	args   interface{}
	meta   interface{}
	loader core.ILoader
	err    error
}

// 创建一个查询配置
func NewQueryConfig() *QueryConfig { return &QueryConfig{} }

// 设置参数, 同 query.WithArgs
func (m *QueryConfig) Args(args interface{}) *QueryConfig {
	m.args = args
	return m
}

// 设置元数据, 同 query.WithMeta
func (m *QueryConfig) Meta(meta interface{}) *QueryConfig {
	m.meta = meta
	return m
}

// 设置加载器, 同 query.WithLoader
func (m *QueryConfig) Loader(loader core.ILoader) *QueryConfig {
	m.loader = loader
	return m
}

// 设置加载函数, 等效于设置加载器
func (m *QueryConfig) LoaderFn(fn loader.LoaderFn, opts ...loader.Option) *QueryConfig {
	m.loader = loader.NewLoader(fn, opts...)
	return m
}

// 获取错误, 查询出错时还可以在这里获取到错误信息
func (m *QueryConfig) GetErr() error {
	return m.err
}

func (m *QueryConfig) setError(err error) {
	m.err = err
}

// 创建一个查询
func NewQuery(bucket string, queryConfig ...*QueryConfig) core.IQuery {
	if len(queryConfig) > 0 {
		qc := queryConfig[0]
		return query.NewQuery(bucket,
			query.WithArgs(qc.args),
			query.WithMeta(qc.meta),
			query.WithLoader(qc.loader),
		)
	}
	return query.NewQuery(bucket)
}

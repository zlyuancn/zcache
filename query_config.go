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
	bucket string
	args   interface{}
	meta   interface{}
	loader core.ILoader
}

// 创建一个查询配置
func NewQueryConfig() *QueryConfig { return &QueryConfig{} }

// 设置 bucket
func (m *QueryConfig) Bucket(bucket string) *QueryConfig {
	m.bucket = bucket
	return m
}

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

// 构建 query
func (m *QueryConfig) Make() core.IQuery {
	return NewQuery(m.bucket,
		query.WithArgs(m.args),
		query.WithMeta(m.meta),
		query.WithLoader(m.loader),
	)
}

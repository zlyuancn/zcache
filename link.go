/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package zcache

import (
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	"github.com/zlyuancn/zcache/loader"
	"github.com/zlyuancn/zcache/query"
)

var (
	// 创建一个加载器
	NewLoader = loader.NewLoader
	// 设置加载器的数据过期时间
	WithLoaderExpire = loader.WithExpire
)

var (
	// 根据选项创建一个查询
	NewQueryWithOption = query.NewQuery
	// 设置查询参数
	WithQueryArgs = query.WithArgs
	// 设置查询元数据
	WithQueryMeta = query.WithMeta
	// 设置查询加载器, 无数据时优先使用这个加载器
	WithQueryLoader = query.WithLoader
	// 设置查询加载函数, 效果等同于设置查询加载器
	WithQueryLoaderFn = func(fn loader.LoaderFn, opts ...loader.Option) query.Option {
		return query.WithLoader(loader.NewLoader(fn, opts...))
	}
)

var (
	// 未找到加载器
	LoaderNotFound = errs.LoaderNotFound
	// 数据为nil
	DataIsNil = errs.DataIsNil
)

// 错误列表
type Errors = errs.Errors

// 尝试将err解包为*Errors
var DecodeErrors = errs.DecodeErrors

type (
	ILoader = core.ILoader
	IQuery  = core.IQuery
)

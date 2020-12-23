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

// 创建一个加载器
var NewLoader = loader.NewLoader

var (
	NewQuery          = query.NewQuery // 创建一个查询
	WithQueryArgs     = query.WithArgs
	WithQueryArgsText = query.WithArgsText
	WithQueryMeta     = query.WithMeta
)

var (
	// 未找到加载器
	LoaderNotFound = errs.LoaderNotFound
	// 数据为nil
	DataIsNil = errs.DataIsNil
)

type (
	IQuery = core.IQuery
)

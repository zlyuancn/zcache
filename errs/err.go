/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package errs

import (
	"errors"
)

// 未找到加载器
var LoaderNotFound = errors.New("loader not found")

// 查询缓存不存在应该返回这个错误
var CacheMiss = errors.New("cache miss")

// 数据为nil
var DataIsNil = errors.New("data is nil")

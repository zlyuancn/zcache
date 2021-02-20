/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/20
   Description :
-------------------------------------------------
*/

package main

import (
	"fmt"

	"github.com/zlyuancn/zcache"
)

func main() {
	cache := zcache.NewCache()
	cache.RegisterLoader("test", "key", zcache.NewLoader(func(query zcache.IQuery) (interface{}, error) {
		return "hello" + query.Args(), nil // 返回 hello + 查询的参数
	}))

	// 提供多个请求参数进行批量获取
	var results []string
	_ = cache.MGet([]zcache.IQuery{
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world1")),
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world2")),
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world3")),
	}, &results)

	fmt.Println(results)
}

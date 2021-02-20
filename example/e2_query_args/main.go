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

	var a string
	_ = cache.Query("test", "key", &a, zcache.WithQueryArgs("world")) // 加入查询参数

	fmt.Println(a)
}

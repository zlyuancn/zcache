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
	"github.com/zlyuancn/zcache/core"
)

func main() {
	cache := zcache.NewCache()

	var a string
	_ = cache.Query("test", &a,
		zcache.NewQueryConfig().LoaderFn(func(query core.IQuery) (interface{}, error) {
			fmt.Println("args", query.Args())      // 打印原始参数
			return "hello" + query.ArgsText(), nil // 返回 hello + 查询的参数文本
		}).Args("world"), // 加入查询参数
	)

	fmt.Println("result", a)
}

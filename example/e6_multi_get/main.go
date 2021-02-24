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
	cache.RegisterLoaderFn("test", "key", func(query zcache.IQuery) (interface{}, error) {
		fmt.Println("重新加载", query.Args())
		return "hello" + query.ArgsText(), nil // 返回 hello + 查询的参数
	})

	var results []string // 批量获取结果的接收变量必须是切片或等同于请求数量的数组

	// 提供多个请求参数进行批量获取, 如果有重复的query我们会进行优化
	_ = cache.MGet([]zcache.IQuery{
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world1")),
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world2")),
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world3")),
		// 这里出现了重复的query, 我们在从缓存或加载器加载数据时会过滤掉这个query, 然后在返回数据给调用者时会将它按顺序返回
		zcache.NewQuery("test", "key", zcache.WithQueryArgs("world1")),
	}, &results)

	fmt.Println(results)
}

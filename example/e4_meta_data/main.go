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
		fmt.Println(query.GlobalId(), "重新加载", query.Meta())
		return "hello", nil // 返回 hello + 查询的参数
	})

	// 创建三个请求参数
	q1 := zcache.NewQuery("test", "key", zcache.WithQueryMeta("world"))
	q2 := zcache.NewQuery("test", "key", zcache.WithQueryMeta("world2"))
	q3 := zcache.NewQueryConfigNK("test", "key").Meta("world3").Make() // 等效于 NewQuery

	// 元数据不会参与 GlobalId 计算, 所以后两次请求是从缓存获取的而不是从加载器重新加载.
	fmt.Println("q1.GlobalId", q1.GlobalId())
	fmt.Println("q2.GlobalId", q2.GlobalId())
	fmt.Println("q3.GlobalId", q3.GlobalId())

	var a string
	_ = cache.Get(q1, &a)
	fmt.Println("q1.Value", a)
	_ = cache.Get(q2, &a)
	fmt.Println("q2.Value", a)
	_ = cache.Get(q3, &a)
	fmt.Println("q3.Value", a)
}

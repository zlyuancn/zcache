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

func exampleA() {
	cache := zcache.NewCache()

	// 注册加载器, 只有相同的 namespace 和 key 才会在加载数据时使用这个加载器
	cache.RegisterLoader("test", zcache.NewLoader(func(query zcache.IQuery) (interface{}, error) {
		// 在这里写入你的db逻辑
		return "hello", nil
	}))

	var a string
	_ = cache.Query("test", &a) // 获取数据, 接收变量必须传入指针

	fmt.Println(a)
}

func exampleB() {
	cache := zcache.NewCache()

	// 注册加载函数, 效果等同于注册加载器
	cache.RegisterLoaderFn("test", func(query zcache.IQuery) (interface{}, error) {
		return "hello", nil
	})

	var a string
	_ = cache.Query("test", &a)

	fmt.Println(a)
}

func main() {
	exampleA()
	exampleB()
}

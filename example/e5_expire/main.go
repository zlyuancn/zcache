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
	"time"

	"github.com/zlyuancn/zcache"
	"github.com/zlyuancn/zcache/loader"
)

func main() {
	cache := zcache.NewCache() // 也可以在这里使用 zcache.WithDefaultExpire 设置全局数据缓存过期时间

	cache.RegisterLoaderFn("test", func(query zcache.IQuery) (interface{}, error) {
		fmt.Println("重新加载")
		return "hello", nil
	}, loader.WithExpire(time.Millisecond*100)) // 设置加载器的数据缓存过期时间

	var a string
	_ = cache.Query("test", &a) // 首次获取由于数据不存在会从加载器获取数据
	fmt.Println(a)
	_ = cache.Query("test", &a) // 这里由于缓存存在会从缓存获取数据
	fmt.Println(a)

	time.Sleep(time.Millisecond * 200) // 等待数据过期

	_ = cache.Query("test", &a) // 这里由于数据缓存过期会重新从加载器获取数据
	fmt.Println(a)
}

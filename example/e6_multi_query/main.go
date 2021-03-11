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
	cache.RegisterLoaderFn("test", func(query zcache.IQuery) (interface{}, error) {
		fmt.Println("重新加载", query.Args())
		return "hello" + query.ArgsText(), nil // 返回 hello + 查询的参数
	})

	var results []string // 批量获取结果的接收变量必须是长度为0的切片或长度等于请求数的数组

	// 提供多个请求参数进行批量获取, 如果有重复的query我们会进行优化
	_ = cache.MQuery("test", &results, // 保存结果的变量必须是指针
		zcache.QC().Args("world1"),
		zcache.QC().Args("world2"),
		zcache.QC().Args("world3"),
		// 这里出现了重复的query, 我们在从缓存或加载器加载数据时会过滤掉这个query, 然后在返回数据给调用者时会将它按顺序返回
		zcache.QC().Args("world1"),
		zcache.QC().Args("world2"),
	)

	fmt.Println(results)
}

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

func exampleA() {
	cache := zcache.NewCache()

	var a string
	_ = cache.Query("test", &a, // 获取数据, 保存结果的变量必须是指针
		// 为query设置查询加载函数, 缓存未命中时执行这个加载函数生成数据, 生成的数据会自动存入缓存
		zcache.QC().LoaderFn(func(query core.IQuery) (interface{}, error) {
			// 在这里写入你的db逻辑
			return "helloA", nil
		}),
	)

	fmt.Println(a)
}

func exampleB() {
	cache := zcache.NewCache()

	var a string
	_ = cache.Query("test", &a,
		// 为query设置查询加载器, 设置查询加载函数的效果等同于设置查询加载器
		zcache.QC().Loader(zcache.NewLoader(func(query core.IQuery) (interface{}, error) {
			return "helloB", nil
		})),
	)

	fmt.Println(a)
}

func main() {
	exampleA()
	exampleB()
}

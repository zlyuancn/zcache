/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package core

import (
	"time"
)

// 加载器
type ILoader interface {
	// 加载数据
	//
	// 缓存数据不存在时会调用此方法获取数据, 获取的数据会自动缓存.
	// 可以在这个过程中加锁防止缓存击穿
	Load(query IQuery) (interface{}, error)
	// 数据过期时间, 数据缓存时会调用这个方法获取缓存时间
	//
	// 可以在这里设置随机有效时间防止缓存雪崩
	Expire() (ex time.Duration)
}

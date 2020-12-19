/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package single_flight

import (
	"github.com/zlyuancn/zcache/core"
)

type noSingleFlight struct{}

// 一个关闭并发查询控制的ISingleFlight
func NoSingleFlight() core.ISingleFlight {
	return new(noSingleFlight)
}

func (*noSingleFlight) Do(globalId uint64, fn func() ([]byte, error)) ([]byte, error) {
	return fn()
}

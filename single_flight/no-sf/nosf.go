/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package no_sf

import (
	"github.com/zlyuancn/zcache/core"
)

type noSingleFlight struct{}

// 一个关闭并发查询控制的ISingleFlight
func NoSingleFlight() core.ISingleFlight {
	return new(noSingleFlight)
}

func (*noSingleFlight) Do(query core.IQuery, fn func(core.IQuery) ([]byte, error)) ([]byte, error) {
	return fn(query)
}

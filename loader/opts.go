/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package loader

import (
	"time"

	"github.com/zlyuancn/zcache/core"
	no_sf "github.com/zlyuancn/zcache/single_flight/no-sf"
)

type Option func(l *Loader)

// 设置过期时间
//
// 如果 endEx > 0 且 ex > 0, 则过期时间在 [ex, endEx-1] 区间随机
// 如果 ex < 0, 则永不过期
// 如果 ex = 0(默认), 则使用全局默认过期时间
func WithExpire(ex time.Duration, endEx ...time.Duration) Option {
	return func(l *Loader) {
		l.startEx, l.endEx = ex, 0
		if len(endEx) > 0 {
			l.endEx = endEx[0]
		}
	}
}

// 设置单跑模块
func WithSingleFlight(sf core.ISingleFlight) Option {
	return func(l *Loader) {
		if sf == nil {
			sf = no_sf.NoSingleFlight()
		}
		l.sf = sf
	}
}

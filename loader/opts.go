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
)

type Option func(l *Loader)

// 设置数据过期时间, 优先级低于传入的 expire
//
// 如果 maxExpire > expire 且 expire > 0, 则过期时间在 [expire, maxExpire-1] 区间随机
// 如果 expire < 0, 则永不过期
// 如果 expire = 0(默认), 则使用全局默认过期时间
func WithExpire(expire time.Duration, maxExpire ...time.Duration) Option {
	return func(l *Loader) {
		l.expire, l.maxExpire = expire, 0
		if len(maxExpire) > 0 {
			l.maxExpire = maxExpire[0]
		}
	}
}

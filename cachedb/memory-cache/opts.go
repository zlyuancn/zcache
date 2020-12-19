/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package memory_cache

import (
	"time"
)

type Option func(m *memoryCache)

// 设置清除过期key时间间隔
func WithCleanupInterval(d time.Duration) Option {
	return func(m *memoryCache) {
		if d <= 0 {
			d = DefaultCleanupInterval
		}
		m.cleanupInterval = d
	}
}

/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package no_cache

import (
	"time"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
)

var _ core.ICacheDB = (*noCache)(nil)

type noCache struct{}

// 创建一个不会缓存数据的ICacheDB
func NoCache() core.ICacheDB { return new(noCache) }

func (*noCache) Set(core.IQuery, []byte, time.Duration) error { return nil }
func (*noCache) Get(core.IQuery) ([]byte, error)              { return nil, errs.CacheMiss }
func (*noCache) MGet(queries ...core.IQuery) ([][]byte, []error) {
	buffs := make([][]byte, len(queries))
	es := make([]error, len(queries))
	for i := range es {
		es[i] = errs.CacheMiss
	}
	return buffs, es
}

func (*noCache) Del(...core.IQuery) error     { return nil }
func (*noCache) DelNamespace(...string) error { return nil }

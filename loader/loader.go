/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package loader

import (
	"errors"
	"math/rand"
	"time"

	"github.com/zlyuancn/zcache/core"
)

type LoaderFn = func(query core.IQuery) (interface{}, error)

var _ core.ILoader = (*Loader)(nil)

type Loader struct {
	fn                LoaderFn      // 加载函数
	expire, maxExpire time.Duration // 有效时间
}

// 创建一个加载器
func NewLoader(fn LoaderFn, opts ...Option) core.ILoader {
	if fn == nil {
		panic(errors.New("load func of loader is empty"))
	}
	l := &Loader{
		fn:        fn,
		expire:    0,
		maxExpire: 0,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

func (l *Loader) Load(query core.IQuery) (interface{}, error) {
	result, err := l.fn(query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (l *Loader) Expire() (ex time.Duration) {
	if l.maxExpire > l.expire && l.expire > 0 {
		return time.Duration(rand.Int63())%(l.maxExpire-l.expire) + (l.expire)
	}
	return l.expire
}

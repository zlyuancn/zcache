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
	fn             LoaderFn      // 加载函数
	startEx, endEx time.Duration // 有效时间
}

// 创建一个加载器
func NewLoader(fn LoaderFn, opts ...Option) core.ILoader {
	if fn == nil {
		panic(errors.New("load func of loader is empty"))
	}
	l := &Loader{
		fn:      fn,
		startEx: 0,
		endEx:   0,
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
	if l.endEx > 0 && l.startEx > 0 {
		return time.Duration(rand.Int63())%(l.endEx-l.startEx) + (l.startEx)
	}
	return l.startEx
}

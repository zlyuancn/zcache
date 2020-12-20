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
	"fmt"
	"math/rand"
	"time"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	single_sf "github.com/zlyuancn/zcache/single_flight/single-sf"
)

type LoaderFn = func(query core.IQuery) (interface{}, error)

var _ core.ILoader = (*Loader)(nil)

type Loader struct {
	fn             LoaderFn           // 加载函数
	startEx, endEx time.Duration      // 有效时间
	sf             core.ISingleFlight // 单跑模块
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
	if l.sf == nil {
		l.sf = single_sf.NewSingleFlight()
	}
	return l
}

func (l *Loader) Load(query core.IQuery, codec core.ICodec) ([]byte, error) {
	return l.sf.Do(query.GlobalId(), func() ([]byte, error) {
		result, err := l.do(query)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, errs.DataIsNil
		}

		bs, err := codec.Encode(result)
		if err != nil {
			return nil, fmt.Errorf("<%T> is can't encode: %s", result, err)
		}
		return bs, nil
	})
}

func (l *Loader) do(query core.IQuery) (result interface{}, err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		switch v := e.(type) {
		case error:
			err = v
		case string:
			err = errors.New(v)
		default:
			err = errors.New(fmt.Sprint(err))
		}
	}()

	result, err = l.fn(query)
	return
}

func (l *Loader) Expire() (ex time.Duration) {
	if l.endEx > 0 && l.startEx > 0 {
		return time.Duration(rand.Int63())%(l.endEx-l.startEx) + (l.startEx)
	}
	return l.startEx
}

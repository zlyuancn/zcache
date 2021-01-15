/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/15
   Description :
-------------------------------------------------
*/

package wrap_call

import (
	"errors"
	"fmt"
)

func WrapCall(fn func() error) (err error) {
	// 包装执行, 拦截panic
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
			err = errors.New(fmt.Sprint(e))
		}
	}()

	err = fn()
	return
}

/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package logger

import (
	"github.com/zlyuancn/zcache/core"
)

var _ core.ILogger = (*noLog)(nil)

type noLog struct{}

func NoLog() core.ILogger             { return new(noLog) }
func (*noLog) Error(v ...interface{}) {}

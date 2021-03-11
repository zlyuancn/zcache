/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/10
   Description :
-------------------------------------------------
*/

package errs

import (
	"bytes"
	"fmt"
	"io"
)

// 强制要求实现error接口
var _ error = (*Errors)(nil)

// error列表
type Errors struct {
	errs []error
}

// 创建一个error列表
func NewErrors(errs ...error) *Errors {
	return &Errors{errs: errs}
}

// 转为error
func (e *Errors) Err() error {
	for _, err := range e.errs {
		if err != nil {
			return e
		}
	}
	return nil
}

// 返回第一个error
func (e *Errors) FirstErr() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e.errs[0]
}

// 获取所有的err
func (e *Errors) Errs() []error {
	return e.errs
}

// 实现error接口
func (e *Errors) Error() string {
	return e.String()
}

// 添加一些错误
func (e *Errors) AddErr(errs ...error) {
	e.errs = append(e.errs, errs...)
}

func (e *Errors) String() string {
	for _, err := range e.errs {
		if err != nil {
			return err.Error()
		}
	}

	return "<nil>"
}

// 格式化输出
//
// %s 会输出第一个错误的文本
// %v 会输出所有错误的文本
// %+v 会输出所有错误的扩展文本
func (e *Errors) Format(s fmt.State, verb rune) {
	if len(e.errs) == 0 {
		_, _ = io.WriteString(s, "<nil>")
		return
	}

	switch verb {
	case 'v':
		var f string
		if s.Flag('+') {
			f = "   %d: %+v\n"
		} else {
			f = "   %d: %v\n"
		}

		var bs bytes.Buffer
		bs.WriteString("errs.Errors: {\n")
		for i, e := range e.errs {
			bs.WriteString(fmt.Sprintf(f, i, e))
		}
		bs.WriteString("}")

		_, _ = io.WriteString(s, bs.String())
	case 's':
		_, _ = io.WriteString(s, e.errs[0].Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.errs[0])
	}
}

// 尝试将err解包为*Errors
func DecodeErrors(err error) (*Errors, bool) {
	if err == nil {
		return nil, false
	}
	e, ok := err.(*Errors)
	return e, ok
}

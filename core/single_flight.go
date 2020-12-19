/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package core

type ISingleFlight interface {
	Do(globalId uint64, fn func() ([]byte, error)) ([]byte, error)
}

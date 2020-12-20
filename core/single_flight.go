/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package core

type ISingleFlight interface {
	Do(query IQuery, fn func(query IQuery) ([]byte, error)) ([]byte, error)
}

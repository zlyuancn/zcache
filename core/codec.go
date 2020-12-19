/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package core

// 编解码器
type ICodec interface {
	// 编码
	Encode(a interface{}) ([]byte, error)
	// 解码
	Decode(data []byte, a interface{}) error
}

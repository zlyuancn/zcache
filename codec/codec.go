/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/vmihailenco/msgpack"
)

// 默认的编解码器
var DefaultCodec = new(msgPackCodec)

// 编解码器
type ICodec interface {
	// 编码
	Encode(i interface{}) ([]byte, error)
	// 解码
	Decode(data []byte, i interface{}) error
}

// 已注册的编解码器
var (
	// 不进行编解码, 编码解码都直接返回原始数据, 原始数据必须为[]byte或*[]byte或string或
	Byte = new(byteCodec)
	// 使用go内置的json包进行编解码
	Json = new(jsonCodec)
	// 使用第三方包json-iterator进行编解码
	JsonIterator = new(jsonIteratorCodec)
	// MsgPack
	MsgPack = new(msgPackCodec)
	// ProtoBuffer
	ProtoBuffer = new(protoBufferCodec)
)

// 不进行编解码
type byteCodec struct{}

func (*byteCodec) Encode(a interface{}) ([]byte, error) {
	switch data := a.(type) {
	case []byte:
		return data, nil
	case *[]byte:
		return *data, nil
	case string:
		return []byte(data), nil
	case *string:
		return []byte(*data), nil
	}
	return nil, fmt.Errorf("<%T> can't convert to []byte or string", a)
}

func (*byteCodec) Decode(data []byte, a interface{}) error {
	switch p := a.(type) {
	case *[]byte:
		*p = data
		return nil
	case *string:
		*p = string(data)
		return nil
	}
	return fmt.Errorf("<%T> can't convert to *[]byte or *string", a)
}

// 使用go内置的json包进行编解码
type jsonCodec struct{}

func (*jsonCodec) Encode(a interface{}) ([]byte, error) {
	return json.Marshal(a)
}

func (*jsonCodec) Decode(data []byte, a interface{}) error {
	return json.Unmarshal(data, a)
}

// 使用第三方包json-iterator进行编解码
type jsonIteratorCodec struct{}

func (*jsonIteratorCodec) Encode(a interface{}) ([]byte, error) {
	return jsoniter.Marshal(a)
}

func (*jsonIteratorCodec) Decode(data []byte, a interface{}) error {
	return jsoniter.Unmarshal(data, a)
}

// MsgPack编解码器
type msgPackCodec struct{}

func (*msgPackCodec) Encode(a interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseJSONTag(true) // 如果没有 msgpack 标记, 使用 json 标记
	err := enc.Encode(a)
	return buf.Bytes(), err
}

func (*msgPackCodec) Decode(data []byte, a interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.UseJSONTag(true) // 如果没有 msgpack 标记, 使用 json 标记
	err := dec.Decode(a)
	return err
}

// ProtoBuffer编解码器
type protoBufferCodec struct{}

func (*protoBufferCodec) Encode(a interface{}) ([]byte, error) {
	if m, ok := a.(proto.Message); ok {
		return proto.Marshal(m)
	}

	return nil, fmt.Errorf("<%T> can't convert to proto.Message", a)
}

func (*protoBufferCodec) Decode(data []byte, a interface{}) error {
	if m, ok := a.(proto.Message); ok {
		return proto.Unmarshal(data, m)
	}

	return fmt.Errorf("<%T> can't convert to proto.Message", a)
}

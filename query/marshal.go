/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package query

import (
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

func Marshal(a interface{}) ([]byte, error) {
	if a == nil {
		return nil, nil
	}

	switch v := a.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case bool:
		if v {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case int:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(v, 10)), nil
	case float32, float64:
		return []byte(fmt.Sprint(v)), nil
	case encoding.TextMarshaler:
		return v.MarshalText()
	case encoding.BinaryMarshaler:
		return v.MarshalBinary()
	}

	rt := reflect.TypeOf(a)
	rv := reflect.ValueOf(a)
	switch rt.Kind() {
	case reflect.Invalid:
		return nil, nil
	case reflect.Ptr:
		return Marshal(rv.Elem().Interface())
	case reflect.Struct:
		fieldCount := rt.NumField()
		if fieldCount == 0 {
			return nil, nil
		}

		var buff bytes.Buffer
		buff.WriteByte('{')
		for i := 0; i < fieldCount; i++ {
			field := rt.Field(i)
			if field.PkgPath != "" {
				continue
			}

			v, err := Marshal(rv.Field(i).Interface())
			if err != nil {
				return nil, err
			}

			buff.WriteString(field.Name)
			buff.WriteByte('=')
			buff.Write(v)
			if i+1 < fieldCount {
				buff.WriteByte('&')
			}
		}
		buff.WriteByte('}')
		return buff.Bytes(), nil
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(a)
}

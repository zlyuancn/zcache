/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/24
   Description :
-------------------------------------------------
*/

package zcache

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
)

// 批量获取, 同MQueryWithContext
func (c *Cache) MQuery(namespace, key string, a interface{}, config ...*QueryConfig) error {
	return c.MQueryWithContext(nil, namespace, key, a, config...)
}

// 批量获取, a必须是长度为0的切片指针或长度等于请求数的数组指针
//
// QueryConfig 的 namespace 和 key 会被替换为这个方法的 namespace 和 key
// 如果有重复的query我们会进行优化, 在从缓存或加载器加载数据时会过滤掉这个query, 然后在返回数据给调用者时会将它按顺序返回
func (c *Cache) MQueryWithContext(ctx context.Context, namespace, key string, a interface{}, config ...*QueryConfig) error {
	queries := make([]core.IQuery, len(config))
	for i, conf := range config {
		queries[i] = conf.Namespace(namespace).Key(key).Make()
	}
	return c.doWithContext(ctx, func() error {
		return c.mQuery(queries, a)
	})
}

// 批量获取, a必须是长度为0的切片指针或长度等于请求数的数组指针.
//
// 如果有重复的query我们会进行优化, 在从缓存或加载器加载数据时会过滤掉这个query, 然后在返回数据给调用者时会将它按顺序返回
//
// deprecated 使用 MQuery
func (c *Cache) MGet(queries []core.IQuery, a interface{}) error {
	return c.MGetWithContext(nil, queries, a)
}

// 批量获取, a必须是长度为0的切片指针或长度等于请求数的数组指针
//
// deprecated 使用 MQueryWithContext
func (c *Cache) MGetWithContext(ctx context.Context, queries []core.IQuery, a interface{}) error {
	return c.doWithContext(ctx, func() error {
		return c.mQuery(queries, a)
	})
}
func (c *Cache) mQuery(queries []core.IQuery, a interface{}) error {
	realQueries := queries
	if len(realQueries) == 0 {
		return nil
	}

	// 过滤重复的query
	queryMap := make(map[uint64]core.IQuery, len(realQueries))
	for _, q := range realQueries {
		queryMap[q.GlobalId()] = q
	}

	var isFilter bool // 是否进行了过滤

	// 如果有重复的, 必然map和slice的长度不一致
	if len(queryMap) != len(realQueries) {
		isFilter = true
		realQueries = make([]core.IQuery, 0, len(queryMap))
		for _, q := range queryMap {
			realQueries = append(realQueries, q)
		}
	}

	// 批量从缓存获取数据
	buffs, cacheErrs := c.cache.MGet(realQueries...)
	if len(buffs) != len(realQueries) || len(cacheErrs) != len(realQueries) {
		panic("cached result is inconsistent with the number of requests")
	}

	// 遍历检查是否存在错误, 补充未命中的数据
	for i, cacheErr := range cacheErrs {
		if cacheErr == nil {
			continue
		}

		q := realQueries[i]
		if cacheErr != errs.CacheMiss { // 非缓存未命中错误
			if c.directReturnOnCacheFault { // 直接报告错误
				cacheErr = fmt.Errorf("load from cache error. query: %s:%s?%s, err: %s", q.Namespace(), q.Key(), q.ArgsText(), cacheErr)
				return cacheErr
			}
			cacheErr = fmt.Errorf("load from cache error, The data will be fetched from the loader. query: %s:%s?%s, err: %s", q.Namespace(), q.Key(), q.ArgsText(), cacheErr)
			c.log.Error(cacheErr)
		}

		// 从加载器获取数据
		bs, err := c.sf.Do(q, c.load)
		if err != nil {
			return err
		}

		buffs[i] = bs
	}

	// 如果没有进行过滤, 顺序和数量是不变的
	if !isFilter {
		return c.writeBuffsTo(buffs, a)
	}

	// 分发
	idMap := make(map[uint64]int, len(realQueries))
	for index, q := range realQueries {
		idMap[q.GlobalId()] = index
	}
	realBuffs := make([][]byte, len(queries))
	for i, q := range queries {
		realBuffs[i] = buffs[idMap[q.GlobalId()]]
	}

	return c.writeBuffsTo(realBuffs, a)
}

// 将批量获取的数据写入a中
func (c *Cache) writeBuffsTo(buffs [][]byte, a interface{}) error {
	// 检查输出
	rt := reflect.TypeOf(a)
	if rt.Kind() != reflect.Ptr {
		panic(errors.New("A must be a pointer"))
	}
	rt = rt.Elem()
	rv := reflect.ValueOf(a).Elem()

	switch rt.Kind() {
	case reflect.Invalid:
		panic(errors.New("A is invalid, it may not be initialized"))
	case reflect.Slice:
		return c.writeBuffsToSlice(buffs, rt, rv)
	case reflect.Array:
		return c.writeBuffsToArray(buffs, rt, rv)
	default:
		panic(errors.New("A must be a slice pointer of length 0 or an array pointer of length equal to the number of requests"))
	}
}

// 将批量获取的数据写入切片中
func (c *Cache) writeBuffsToSlice(buffs [][]byte, sliceType reflect.Type, sliceValue reflect.Value) (err error) {
	if sliceValue.Kind() == reflect.Invalid {
		panic(errors.New("A is invalid"))
	}
	if sliceValue.Len() != 0 {
		panic(errors.New("length of the slice must be 0"))
	}

	itemType := sliceType.Elem()                // 获取内容类型
	itemIsPtr := itemType.Kind() == reflect.Ptr // 检查内容类型是否为指针
	if itemIsPtr {
		itemType = itemType.Elem() // 获取指针指向的真正的内容类型
	}

	items := make([]reflect.Value, len(buffs))
	for i, bs := range buffs {
		child := reflect.New(itemType) // 创建一个相同类型的指针
		if err = c.unmarshal(bs, child.Interface()); err != nil {
			return err
		}

		if !itemIsPtr {
			child = child.Elem() // 如果想要的不是指针那么获取它的内容
		}
		items[i] = child
	}

	values := reflect.Append(sliceValue, items...) // 构建内容切片
	sliceValue.Set(values)                         // 将内容切片写入原始切片中
	return nil
}

// 将批量获取的数据写入数组中
func (c *Cache) writeBuffsToArray(buffs [][]byte, arrayType reflect.Type, arrayValue reflect.Value) (err error) {
	if arrayValue.Kind() == reflect.Invalid {
		panic(errors.New("A is invalid"))
	}
	if arrayType.Len() != len(buffs) {
		panic(errors.New("array length is not equal to the number of requests"))
	}

	itemType := arrayType.Elem()                // 获取内容类型
	itemIsPtr := itemType.Kind() == reflect.Ptr // 检查内容类型是否为指针
	if itemIsPtr {
		itemType = itemType.Elem() // 获取指针指向的真正的内容类型
	}

	for i, bs := range buffs {
		child := reflect.New(itemType) // 创建一个相同类型的指针
		if err = c.unmarshal(bs, child.Interface()); err != nil {
			return err
		}

		if !itemIsPtr {
			child = child.Elem() // 如果想要的不是指针那么获取它的内容
		}
		arrayValue.Index(i).Set(child)
	}
	return nil
}

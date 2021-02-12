/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package zcache

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/zlyuancn/zcache/cachedb/memory-cache"
	single_sf "github.com/zlyuancn/zcache/single_flight/single-sf"
	"github.com/zlyuancn/zcache/wrap_call"

	"github.com/zlyuancn/zcache/codec"
	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
	"github.com/zlyuancn/zcache/logger"
)

const (
	defaultDirectReturnOnCacheFault = true
	defaultPanicOnLoaderExists      = true
)

type Cache struct {
	cache                    core.ICacheDB // 缓存数据库
	startEx, endEx           time.Duration // 默认过期时间
	directReturnOnCacheFault bool          // 在缓存故障时直接返回

	codec core.ICodec // 编解码器

	loaders             map[uint64]core.ILoader // 加载器配置
	panicOnLoaderExists bool                    // 注册加载器时如果加载器已存在会panic, 设为false会替换旧的加载器
	mx                  sync.RWMutex            // 对注册的加载器加锁
	sf                  core.ISingleFlight      // 单跑模块

	log core.ILogger // 日志
}

func NewCache(opts ...Option) *Cache {
	c := &Cache{
		startEx:                  0,
		endEx:                    0,
		directReturnOnCacheFault: defaultDirectReturnOnCacheFault,

		codec: codec.DefaultCodec,

		loaders:             make(map[uint64]core.ILoader),
		panicOnLoaderExists: defaultPanicOnLoaderExists,
	}

	for _, o := range opts {
		o(c)
	}

	if c.cache == nil {
		c.cache = memory_cache.NewMemoryCache()
	}
	if c.sf == nil {
		c.sf = single_sf.NewSingleFlight()
	}
	if c.log == nil {
		c.log = logger.NoLog()
	}
	return c
}

// 注册加载器
func (c *Cache) RegisterLoader(namespace, key string, loader core.ILoader) {
	if namespace == "" {
		panic(errors.New("namespace is empty"))
	}
	if key == "" {
		panic(errors.New("key is empty"))
	}

	loader_id := c.makeLoaderId(namespace, key)
	c.mx.Lock()
	if c.loaders[loader_id] != nil && c.panicOnLoaderExists {
		c.mx.Unlock()
		panic("loader is exists")
	}
	c.loaders[loader_id] = loader
	c.mx.Unlock()
}

// 获取加载器
func (c *Cache) loader(namespace, key string) core.ILoader {
	loader_id := c.makeLoaderId(namespace, key)
	c.mx.RLock()
	loader := c.loaders[loader_id]
	c.mx.RUnlock()
	return loader
}

// 获取数据
func (c *Cache) Get(query core.IQuery, a interface{}) error {
	return c.GetWithContext(nil, query, a)
}

// 获取数据
func (c *Cache) GetWithContext(ctx context.Context, query core.IQuery, a interface{}) error {
	return c.doWithContext(ctx, func() error {
		return c.get(query, a)
	})
}
func (c *Cache) get(query core.IQuery, a interface{}) error {
	// 从缓存获取数据
	bs, cacheErr := c.cache.Get(query)
	if cacheErr == nil {
		return c.unmarshal(bs, a)
	}
	if cacheErr != errs.CacheMiss { // 非缓存未命中错误
		cacheErr = fmt.Errorf("load from cache error, The data will be fetched from the loader. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.Args(), cacheErr)
		if c.directReturnOnCacheFault { // 直接报告错误
			return cacheErr
		}
		c.log.Error(cacheErr)
	}

	// 从加载器获取数据
	bs, err := c.sf.Do(query, c.load)
	if err != nil {
		return err
	}

	return c.unmarshal(bs, a)
}

// 批量获取, a必须是长度为0的切片指针或长度等于请求数的数组指针
func (c *Cache) MGet(queries []core.IQuery, a interface{}) error {
	return c.MGetWithContext(nil, queries, a)
}

// 批量获取, a必须是长度为0的切片指针或长度等于请求数的数组指针
func (c *Cache) MGetWithContext(ctx context.Context, queries []core.IQuery, a interface{}) error {
	return c.doWithContext(ctx, func() error {
		return c.mGet(queries, a)
	})
}
func (c *Cache) mGet(queries []core.IQuery, a interface{}) error {
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

		query := realQueries[i]
		if cacheErr != errs.CacheMiss { // 非缓存未命中错误
			if c.directReturnOnCacheFault { // 直接报告错误
				cacheErr = fmt.Errorf("load from cache error. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.Args(), cacheErr)
				return cacheErr
			}
			cacheErr = fmt.Errorf("load from cache error, The data will be fetched from the loader. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.Args(), cacheErr)
			c.log.Error(cacheErr)
		}

		// 从加载器获取数据
		bs, err := c.sf.Do(query, c.load)
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

// 加载数据并写入缓存
func (c *Cache) load(query core.IQuery) (bs []byte, err error) {
	err = wrap_call.WrapCall(func() error {
		// 加载数据
		loader := c.loader(query.Namespace(), query.Key())
		if loader == nil {
			return errs.LoaderNotFound
		}
		result, err := loader.Load(query)
		if err != nil {
			return fmt.Errorf("load data error from loader: %s", err)
		}

		// 编码
		bs, err = c.marshal(result)
		if err != nil {
			return err
		}

		// 写入缓存
		cacheErr := c.cache.Set(query, bs, c.makeExpire(loader.Expire()))
		if cacheErr != nil {
			cacheErr = fmt.Errorf("write to cache error. query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.Args(), cacheErr)
			if c.directReturnOnCacheFault {
				return cacheErr
			}
			c.log.Error(cacheErr)
		}
		return nil
	})
	return bs, err
}

// 设置数据到缓存
func (c *Cache) Set(query core.IQuery, a interface{}, ex ...time.Duration) error {
	return c.SetWithContext(nil, query, a, ex...)
}

// 设置数据到缓存
//
// ex < 0 表示永不过期, ex = 0 或未设置表示使用默认过期时间
func (c *Cache) SetWithContext(ctx context.Context, query core.IQuery, a interface{}, ex ...time.Duration) error {
	return c.doWithContext(ctx, func() error {
		bs, err := c.marshal(a)
		if err != nil {
			return err
		}

		err = c.cache.Set(query, bs, c.makeExpire(ex...))
		if err != nil {
			return fmt.Errorf("write to cache error, query: %s:%s?%s, err: %s", query.Namespace(), query.Key(), query.Args(), err)
		}
		return nil
	})
}

// 删除指定数据
func (c *Cache) Del(queries ...core.IQuery) error {
	return c.DelWithContext(nil, queries...)
}

// 删除指定数据
func (c *Cache) DelWithContext(ctx context.Context, queries ...core.IQuery) (err error) {
	if len(queries) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		return c.cache.Del(queries...)
	})
}

// 删除命名空间下所有数据
func (c *Cache) DelNamespace(namespaces ...string) error {
	return c.DelSpaceWithContext(nil, namespaces...)
}

// 删除命名空间下所有数据
func (c *Cache) DelSpaceWithContext(ctx context.Context, namespaces ...string) error {
	if len(namespaces) == 0 {
		return nil
	}
	return c.doWithContext(ctx, func() error {
		return c.cache.DelNamespace(namespaces...)
	})
}

// 将数据解码到a
func (c *Cache) marshal(a interface{}) ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	bs, err := c.codec.Encode(a)
	if err != nil {
		return nil, fmt.Errorf("<%T> is can't encode: %s", a, err)
	}
	return bs, nil
}

// 将数据解码到a
func (c *Cache) unmarshal(bs []byte, a interface{}) error {
	if len(bs) == 0 && c.codec != codec.Byte {
		return errs.DataIsNil
	}
	err := c.codec.Decode(bs, a)
	if err != nil {
		return fmt.Errorf("can't decode to <%T>: %s", a, err)
	}
	return nil
}

// 构建加载器id
func (c *Cache) makeLoaderId(namespace, key string) uint64 {
	f := fnv.New64a()
	_, _ = f.Write([]byte(namespace))
	_, _ = f.Write([]byte{':'})
	_, _ = f.Write([]byte(key))
	return f.Sum64()
}

// 为一个执行添加上下文
func (c *Cache) doWithContext(ctx context.Context, fn func() error) (err error) {
	if ctx == nil || ctx == context.Background() || ctx == context.TODO() {
		return wrap_call.WrapCall(fn)
	}

	done := make(chan struct{}, 1)
	go func() {
		err = wrap_call.WrapCall(fn)
		done <- struct{}{}
	}()

	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// 构建超时
func (c *Cache) makeExpire(ex ...time.Duration) time.Duration {
	if len(ex) > 0 && ex[0] != 0 {
		return ex[0]
	}
	if c.endEx > 0 && c.startEx > 0 {
		return time.Duration(rand.Int63())%(c.endEx-c.startEx) + (c.startEx)
	}
	return c.startEx
}

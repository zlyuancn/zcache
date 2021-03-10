/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package memory_cache

import (
	"sync"
	"time"

	go_cache "github.com/patrickmn/go-cache"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
)

var _ core.ICacheDB = (*memoryCache)(nil)

const (
	NoExpiration           = time.Duration(-1) // 无过期时间
	DefaultCleanupInterval = time.Minute * 5   // 默认清除过期key时间
)

type memoryCache struct {
	buckets map[string]*go_cache.Cache
	mx      sync.RWMutex

	// 每隔一段时间后清理过期的key
	cleanupInterval time.Duration
}

// 创建一个内存缓存
func NewMemoryCache(opts ...Option) core.ICacheDB {
	m := &memoryCache{
		buckets:         make(map[string]*go_cache.Cache),
		cleanupInterval: DefaultCleanupInterval,
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// 获取桶
func (m *memoryCache) bucket(bucket string) *go_cache.Cache {
	m.mx.RLock()
	cache, ok := m.buckets[bucket]
	m.mx.RUnlock()
	if ok {
		return cache
	}

	m.mx.Lock()
	if cache, ok = m.buckets[bucket]; ok {
		m.mx.Unlock()
		return cache
	}

	cache = go_cache.New(0, m.cleanupInterval)
	m.buckets[bucket] = cache
	m.mx.Unlock()
	return cache
}

func (m *memoryCache) Set(query core.IQuery, bs []byte, ex time.Duration) error {
	if ex <= 0 {
		ex = NoExpiration
	}
	m.bucket(query.Bucket()).Set(query.ArgsText(), bs, ex)
	return nil
}
func (m *memoryCache) Get(query core.IQuery) ([]byte, error) {
	v, ok := m.bucket(query.Bucket()).Get(query.ArgsText())
	if !ok {
		return nil, errs.CacheMiss
	}
	if v == nil {
		return nil, nil
	}
	return v.([]byte), nil
}
func (m *memoryCache) MGet(queries ...core.IQuery) ([][]byte, []error) {
	buffs := make([][]byte, len(queries))
	es := make([]error, len(queries))
	for i, query := range queries {
		v, ok := m.bucket(query.Bucket()).Get(query.ArgsText())
		if !ok {
			es[i] = errs.CacheMiss
			continue
		}

		if v != nil {
			buffs[i] = v.([]byte)
		}
	}
	return buffs, es
}

func (m *memoryCache) Del(queries ...core.IQuery) error {
	for _, query := range queries {
		m.bucket(query.Bucket()).Delete(query.ArgsText())
	}
	return nil
}
func (m *memoryCache) DelBucket(buckets ...string) error {
	m.mx.Lock()
	for _, bucket := range buckets {
		delete(m.buckets, bucket)
	}
	m.mx.Unlock()
	return nil
}

func (m *memoryCache) Close() error {
	m.mx.Lock()
	buckets := m.buckets
	m.buckets = make(map[string]*go_cache.Cache)
	m.mx.Unlock()

	for _, bucket := range buckets {
		bucket.Flush()
	}
	return nil
}

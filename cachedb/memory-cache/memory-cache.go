/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package memory_cache

import (
	"bytes"
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
func (m *memoryCache) bucket(namespace string) *go_cache.Cache {
	m.mx.RLock()
	c, ok := m.buckets[namespace]
	m.mx.RUnlock()
	if ok {
		return c
	}

	m.mx.Lock()
	if c, ok = m.buckets[namespace]; ok {
		m.mx.Unlock()
		return c
	}

	c = go_cache.New(0, m.cleanupInterval)
	m.buckets[namespace] = c
	m.mx.Unlock()
	return c
}

func (m *memoryCache) Set(query core.IQuery, bs []byte, ex time.Duration) error {
	if ex <= 0 {
		ex = NoExpiration
	}
	m.bucket(query.Namespace()).Set(m.makeKey(query), bs, ex)
	return nil
}
func (m *memoryCache) Get(query core.IQuery) ([]byte, error) {
	v, ok := m.bucket(query.Namespace()).Get(m.makeKey(query))
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
		v, ok := m.bucket(query.Namespace()).Get(m.makeKey(query))
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
		m.bucket(query.Namespace()).Delete(m.makeKey(query))
	}
	return nil
}
func (m *memoryCache) DelNamespace(namespaces ...string) error {
	m.mx.Lock()
	for _, namespace := range namespaces {
		delete(m.buckets, namespace)
	}
	m.mx.Unlock()
	return nil
}

func (m *memoryCache) makeKey(query core.IQuery) string {
	if query.Args() == "" {
		return query.Key()
	}

	var buff bytes.Buffer
	buff.WriteString(query.Key())
	buff.WriteByte('?')
	buff.WriteString(query.Args())
	return buff.String()
}

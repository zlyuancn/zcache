/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	rredis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/zlyuancn/zcache"
	"github.com/zlyuancn/zcache/cachedb/memory-cache"
	redis_cache "github.com/zlyuancn/zcache/cachedb/redis-cache"
	"github.com/zlyuancn/zcache/codec"
)

func makeMemoryCache() *zcache.Cache {
	return zcache.NewCache(
		zcache.WithCacheDB(memory_cache.NewMemoryCache()),
		zcache.WithCodec(codec.Byte),
	)
}
func makeRedisCache() *zcache.Cache {
	client := rredis.NewClient(&rredis.Options{
		Addr:        "127.0.0.1:6379",
		Password:    "",
		DB:          0,
		PoolSize:    50,
		DialTimeout: time.Second * 3,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return zcache.NewCache(
		zcache.WithCacheDB(redis_cache.NewRedisCache(client)),
		zcache.WithCodec(codec.Byte),
	)
}

func TestMemoryCacheGet(t *testing.T) {
	cache := makeMemoryCache()
	testMemoryCacheGet(t, cache)
}
func TestMemoryCacheSet(t *testing.T) {
	cache := makeMemoryCache()
	testMemoryCacheSet(t, cache)
}
func TestMemoryCacheDel(t *testing.T) {
	cache := makeMemoryCache()
	testMemoryCacheDel(t, cache)
}
func TestMemoryCacheDelNamespace(t *testing.T) {
	cache := makeMemoryCache()
	testMemoryCacheDelNamespace(t, cache)
}

func TestRedisCacheGet(t *testing.T) {
	cache := makeRedisCache()
	testMemoryCacheGet(t, cache)
}
func TestRedisCacheSet(t *testing.T) {
	cache := makeRedisCache()
	testMemoryCacheSet(t, cache)
}
func TestRedisCacheDel(t *testing.T) {
	cache := makeRedisCache()
	testMemoryCacheDel(t, cache)
}
func TestRedisCacheDelNamespace(t *testing.T) {
	cache := makeRedisCache()
	testMemoryCacheDelNamespace(t, cache)
}

func testMemoryCacheGet(t *testing.T, cache *zcache.Cache) {
	const namespace = "test"
	const key = "key"
	loader := zcache.NewLoader(func(query zcache.IQuery) (i interface{}, err error) {
		s := fmt.Sprintf("%s:%s?%d", query.Namespace(), query.Key(), query.GlobalId())
		return s, nil
	})

	cache.RegisterLoader(namespace, key, loader)

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs([]interface{}{k, v}))
		expect := fmt.Sprintf("%s:%s?%d", q.Namespace(), q.Key(), q.GlobalId())

		var result string
		err := cache.Get(q, &result)
		require.NoError(t, err)
		require.Equal(t, result, expect)
	}
}
func testMemoryCacheSet(t *testing.T, cache *zcache.Cache) {
	const namespace = "test"
	const key = "key"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs([]interface{}{k, v}))
		expect := fmt.Sprintf("%s:%s?%d", q.Namespace(), q.Key(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err)

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err)
		require.Equal(t, result, expect)
	}
}
func testMemoryCacheDel(t *testing.T, cache *zcache.Cache) {
	const namespace = "test"
	const key = "key"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs([]interface{}{k, v}))
		expect := fmt.Sprintf("%s:%s?%d", q.Namespace(), q.Key(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err)

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err)
		require.Equal(t, result, expect)

		err = cache.Del(q)
		require.NoError(t, err)

		result = ""
		err = cache.Get(q, &result)
		require.Equal(t, err, zcache.LoaderNotFound)
	}
}
func testMemoryCacheDelNamespace(t *testing.T, cache *zcache.Cache) {
	const namespace = "test"
	const key = "key"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs([]interface{}{k, v}))
		expect := fmt.Sprintf("%s:%s?%d", q.Namespace(), q.Key(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err)

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err)
		require.Equal(t, result, expect)

		err = cache.Del(q)
		require.NoError(t, err)
	}

	err := cache.DelNamespace(namespace)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs([]interface{}{k, v}))

		var result string
		err = cache.Get(q, &result)
		require.Equal(t, err, zcache.LoaderNotFound)
	}
}

func BenchmarkMemoryCache_10k(b *testing.B) {
	benchmarkAny(b, makeMemoryCache(), 1e4)
}

func BenchmarkRedisCache_10k(b *testing.B) {
	benchmarkAny(b, makeRedisCache(), 1e4)
}

func benchmarkAny(b *testing.B, cache *zcache.Cache, maxKeyCount int) {
	rand.Seed(time.Now().UnixNano())

	const byteLen = 512
	const namespace = "benchmark"
	const key = "key"

	expects := make([][]byte, maxKeyCount)
	for i := 0; i < maxKeyCount; i++ {
		bs := make([]byte, byteLen)
		for j := 0; j < byteLen; j++ {
			bs[j] = byte(rand.Int() % 256)
		}
		expects[i] = bs

		q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs(i))
		err := cache.Set(q, bs)
		require.NoError(b, err, "数据设置失败")
	}

	// 缓存随机key, 因为它本身速度太慢了(117ns/op)
	randKeys := make([]int, 1<<20)
	for i := 0; i < len(randKeys); i++ {
		randKeys[i] = rand.Int() % maxKeyCount
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		i := 0
		for p.Next() {
			i++
			k := randKeys[i&(len(randKeys)-1)]
			q := zcache.NewQuery(namespace, key, zcache.WithQueryArgs(k))

			var bs []byte
			err := cache.Get(q, &bs)
			if err != nil {
				b.Fatalf("数据加载失败: %s", err)
			}
			if len(bs) != byteLen {
				b.Fatalf("数据长度不一致, need %d, got %d", byteLen, len(bs))
			}
			if !bytes.Equal(bs, expects[k]) {
				b.Fatalf("数据不一致: key: %d", k)
			}
		}
	})
}

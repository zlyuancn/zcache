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

func TestMemoryCache(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		cache := makeMemoryCache()
		testCacheGet(t, cache)
	})
	t.Run("Set", func(t *testing.T) {
		cache := makeMemoryCache()
		testCacheSet(t, cache)
	})
	t.Run("Del", func(t *testing.T) {
		cache := makeMemoryCache()
		testCacheDel(t, cache)
	})
	t.Run("DelBucket", func(t *testing.T) {
		cache := makeMemoryCache()
		testCacheDelBucket(t, cache)
	})
	t.Run("Expire", func(t *testing.T) {
		cache := makeMemoryCache()
		testCacheExpire(t, cache)
	})
}

func TestRedisCacheGet(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		cache := makeRedisCache()
		testCacheGet(t, cache)
	})
	t.Run("Set", func(t *testing.T) {
		cache := makeRedisCache()
		testCacheSet(t, cache)
	})
	t.Run("Del", func(t *testing.T) {
		cache := makeRedisCache()
		testCacheDel(t, cache)
	})
	t.Run("DelBucket", func(t *testing.T) {
		cache := makeRedisCache()
		testCacheDelBucket(t, cache)
	})
	t.Run("Expire", func(t *testing.T) {
		cache := makeRedisCache()
		testCacheExpire(t, cache)
	})
}

func testCacheGet(t *testing.T, cache *zcache.Cache) {
	const bucket = "test"
	cache.RegisterLoaderFn(bucket, func(query zcache.IQuery) (i interface{}, err error) {
		s := fmt.Sprintf("%s?%d", query.ArgsText(), query.GlobalId())
		return s, nil
	})

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.Q(bucket, zcache.QC().Args([]interface{}{k, v}))
		expect := fmt.Sprintf("%s?%d", q.ArgsText(), q.GlobalId())

		var result string
		err := cache.Get(q, &result)
		require.NoError(t, err, "获取失败")
		require.Equal(t, result, expect, "数据和预期不符")
	}
}
func testCacheSet(t *testing.T, cache *zcache.Cache) {
	const bucket = "test"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.Q(bucket, zcache.QC().Args([]interface{}{k, v}))
		expect := fmt.Sprintf("%s?%d", q.ArgsText(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err, "存入失败")

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err, "获取失败")
		require.Equal(t, result, expect, "数据和预期不符")
	}
}
func testCacheDel(t *testing.T, cache *zcache.Cache) {
	const bucket = "test"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.Q(bucket, zcache.QC().Args([]interface{}{k, v}))
		expect := fmt.Sprintf("%s?%d", q.ArgsText(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err, "存入失败")

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err, "获取失败")
		require.Equal(t, result, expect, "数据和预期不符")

		err = cache.Remove(q)
		require.NoError(t, err, "移除失败")

		err = cache.Get(q, &result)
		require.Equal(t, err, zcache.LoaderNotFound, "数据和预期不符")
	}
}
func testCacheDelBucket(t *testing.T, cache *zcache.Cache) {
	const bucket = "test"

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.Q(bucket, zcache.QC().Args([]interface{}{k, v}))
		expect := fmt.Sprintf("%s?%d", q.ArgsText(), q.GlobalId())
		err := cache.Set(q, expect)
		require.NoError(t, err)

		var result string
		err = cache.Get(q, &result)
		require.NoError(t, err)
		require.Equal(t, result, expect)
	}

	err := cache.DelBucket(bucket)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("k%d", i)
		v := fmt.Sprintf("v%d", i)

		q := zcache.Q(bucket, zcache.QC().Args([]interface{}{k, v}))

		var result string
		err = cache.Get(q, &result)
		require.Equal(t, err, zcache.LoaderNotFound)
	}
}
func testCacheExpire(t *testing.T, cache *zcache.Cache) {
	const bucket = "test"
	const expect = "hello"
	err := cache.Save(bucket, expect, time.Millisecond*100)
	require.NoError(t, err, "存入失败")

	var result string
	err = cache.Query(bucket, &result)
	require.NoError(t, err, "获取失败")
	require.Equal(t, result, expect, "数据和预期不符")

	time.Sleep(time.Millisecond*200)

	err = cache.Query(bucket, &result)
	require.Equal(t, err, zcache.LoaderNotFound, "数据和预期不符")
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
	const bucket = "benchmark"

	expects := make([][]byte, maxKeyCount)
	for i := 0; i < maxKeyCount; i++ {
		bs := make([]byte, byteLen)
		for j := 0; j < byteLen; j++ {
			bs[j] = byte(rand.Int() % 256)
		}
		expects[i] = bs

		err := cache.Save(bucket, bs, 0, zcache.QC().Args(i))
		require.NoError(b, err, "数据设置失败")
	}

	// 缓存随机key
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

			var bs []byte
			err := cache.Query(bucket, &bs, zcache.QC().Args(k))
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

/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	rredis "github.com/go-redis/redis/v8"

	"github.com/zlyuancn/zcache"
	memory_cache "github.com/zlyuancn/zcache/cachedb/memory-cache"
	redis_cache "github.com/zlyuancn/zcache/cachedb/redis-cache"
)

func makeMemoryCache() *zcache.Cache {
	return zcache.NewCache(
		zcache.WithCacheDB(memory_cache.NewMemoryCache()),
	)
}
func makeRedisCache(host string, pwd string, db int) *zcache.Cache {
	client := rredis.NewClient(&rredis.Options{
		Addr:        host,
		Password:    pwd,
		DB:          db,
		PoolSize:    50,
		DialTimeout: time.Second * 3,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	return zcache.NewCache(
		zcache.WithCacheDB(redis_cache.NewRedisCache(client)),
	)
}

func benchmark_any(cache *zcache.Cache, maxKeyCount int, clientCount int) {
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

		q := zcache.NewQuery(bucket, zcache.WithQueryArgs(i))
		err := cache.Set(q, bs)
		if err != nil {
			log.Fatalf("数据设置失败: %s", err)
		}
	}

	log.Print("开始")

	done := make(chan struct{})
	tn := time.NewTicker(time.Second)

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-interrupt
		close(done)
	}()

	for {
		select {
		case <-done:
			tn.Stop()
			log.Print("结束")
			return
		case <-tn.C:
			for j := 0; j < clientCount; j++ {
				go func() {
					rn := rand.Int()
					time.Sleep(time.Duration(rn&1023+1) * time.Millisecond) // 随机等待1秒内
					args := rn % maxKeyCount
					q := zcache.NewQuery(bucket, zcache.WithQueryArgs(args))

					var bs []byte
					err := cache.Get(q, &bs)
					if err != nil {
						log.Fatalf("数据加载失败: %s", err)
					}
					if len(bs) != byteLen {
						log.Fatalf("数据长度不一致, need %d, got %d", byteLen, len(bs))
					}
					if !bytes.Equal(bs, expects[args]) {
						log.Fatalf("数据不一致: key: %d", args)
					}
				}()
			}
		}
	}
}

func main() {
	cacheType := flag.String("cache", "memory", "cache类型: memory(默认), redis")
	keyCount := flag.Int("key_count", 1000, "key数量, 默认1000")
	clientCount := flag.Int("client_count", 1000, "客户端数量, 默认1000")
	redis_host := flag.String("redis_host", "127.0.0.1:6379", "redis地址, 默认127.0.0.1:6379")
	redis_pwd := flag.String("redis_pwd", "", "redis密码")
	redis_db := flag.Int("redis_db", 0, "redis的db, 默认0")
	flag.Parse()

	var cache *zcache.Cache
	switch *cacheType {
	case "memory":
		cache = makeMemoryCache()
	case "redis":
		cache = makeRedisCache(*redis_host, *redis_pwd, *redis_db)
	default:
		log.Fatal("cache类型为 memory(默认), redis")
	}
	benchmark_any(cache, *keyCount, *clientCount)
}

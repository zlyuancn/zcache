/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package redis_cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	rredis "github.com/go-redis/redis/v8"

	"github.com/zlyuancn/zcache/core"
	"github.com/zlyuancn/zcache/errs"
)

// 默认参数分隔符
const defaultArgsSep = ":"

var _ core.ICacheDB = (*redisCache)(nil)

type redisCache struct {
	client    rredis.UniversalClient // redis客户端
	keyPrefix string                 // key前缀
	argsSep   string
}

func NewRedisCache(redisClient rredis.UniversalClient, opts ...Option) core.ICacheDB {
	r := &redisCache{
		client:  redisClient,
		argsSep: defaultArgsSep,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *redisCache) Set(query core.IQuery, bs []byte, ex time.Duration) error {
	if ex <= 0 {
		ex = -1
	}
	return r.client.Set(context.Background(), r.makeKey(query), bs, ex).Err()
}
func (r *redisCache) Get(query core.IQuery) ([]byte, error) {
	result, err := r.client.Get(context.Background(), r.makeKey(query)).Bytes()
	if err == rredis.Nil {
		return nil, errs.CacheMiss
	}
	return result, err
}
func (r *redisCache) MGet(queries ...core.IQuery) ([][]byte, []error) {
	buffs := make([][]byte, len(queries))
	es := make([]error, len(queries))

	// 构建key
	keys := make([]string, len(queries))
	for i, query := range queries {
		keys[i] = r.makeKey(query)
	}

	// 查询数据
	results, err := r.client.MGet(context.Background(), keys...).Result()
	if err == nil && len(results) != len(queries) { // 获取到数据, 但是数量不对
		err = errors.New("cached result is inconsistent with the number of requests")
	}

	// MGet不会出现 redis.Nil 错误, 但是我们要考虑.
	// 这里不需要处理 redis.Nil 错误, 下面的循环会检查缓存未命中
	if err != nil && err != rredis.Nil {
		for i := range es { // 一旦有错误, 所有值的错误都一样
			es[i] = err
		}
		return buffs, es
	}

	// 循环检查所有结果
	for i, result := range results {
		switch v := result.(type) {
		case nil: // nil 一定是缓存未命中
			es[i] = errs.CacheMiss
		case string:
			buffs[i] = []byte(v)
		case []byte:
			buffs[i] = v
		default: // 虽然不会出现, 但是我们要处理
			es[i] = fmt.Errorf("Unrecognized redis result type <%T>", result)
		}
	}
	return buffs, es
}

func (r *redisCache) Del(queries ...core.IQuery) error {
	keys := make([]string, len(queries))
	for i, query := range queries {
		keys[i] = r.makeKey(query)
	}

	err := r.client.Del(context.Background(), keys...).Err()
	if err == rredis.Nil { // 虽然测试了不会出现 redis.Nil, 但是我们要考虑
		return nil
	}
	return err
}

func (r *redisCache) DelBucket(buckets ...string) error {
	for _, bucket := range buckets {
		key := r.keyPrefix + bucket + defaultArgsSep + "*"
		if err := r.scanDelKey(key); err != nil {
			return err
		}
	}
	return nil
}

func (r *redisCache) makeKey(query core.IQuery) string {
	var buff bytes.Buffer
	if r.keyPrefix != "" {
		buff.WriteString(r.keyPrefix)
	}
	buff.WriteString(query.Bucket())
	if query.ArgsText() != "" {
		buff.WriteString(r.argsSep)
		buff.WriteString(query.ArgsText())
	}
	return buff.String()
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

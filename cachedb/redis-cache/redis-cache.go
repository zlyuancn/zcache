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

var _ core.ICacheDB = (*redisCache)(nil)

type redisCache struct {
	client    rredis.UniversalClient // redis客户端
	keyPrefix string                 // key前缀
}

func NewRedisCache(redisClient rredis.UniversalClient, opts ...Option) core.ICacheDB {
	r := &redisCache{
		client: redisClient,
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

	keys := make([]string, len(queries))
	for i, query := range queries {
		keys[i] = r.makeKey(query)
	}
	results, err := r.client.MGet(context.Background(), keys...).Result()
	if err == nil && len(results) != len(queries) {
		err = errors.New("cached result is inconsistent with the number of requests")
	}
	if err != nil {
		for i := range es {
			es[i] = err
		}
		return buffs, es
	}

	for i, result := range results {
		switch v := result.(type) {
		case nil:
			es[i] = errs.CacheMiss
		case string:
			buffs[i] = []byte(v)
		case []byte:
			buffs[i] = v
		default:
			es[i] = fmt.Errorf("Unrecognized result type <%T>", result)
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
	if err == rredis.Nil {
		return nil
	}
	return err
}

func (r *redisCache) DelNamespace(namespaces ...string) error {
	for _, namespace := range namespaces {
		key := r.keyPrefix + namespace + ":*"
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
	buff.WriteString(query.Namespace())
	buff.WriteByte(':')
	buff.WriteString(query.Key())
	if query.ArgsText() != "" {
		buff.WriteByte(':')
		buff.WriteString(query.ArgsText())
	}
	return buff.String()
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

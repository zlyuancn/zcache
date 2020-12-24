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
	"errors"
	"time"

	rredis "github.com/go-redis/redis"

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
	return r.client.Set(r.makeKey(query), bs, ex).Err()
}
func (r *redisCache) Get(query core.IQuery) ([]byte, error) {
	result, err := r.client.Get(r.makeKey(query)).Bytes()
	if err == rredis.Nil {
		return nil, errs.CacheMiss
	}
	return result, err
}

func (r *redisCache) Del(query core.IQuery) error {
	err := r.client.Del(r.makeKey(query)).Err()
	if err == rredis.Nil {
		return nil
	}
	return err
}
func (r *redisCache) DelNamespace(namespace string) error {
	return errors.New("del namespace is un implement")
}

func (r *redisCache) makeKey(query core.IQuery) string {
	var buff bytes.Buffer
	if r.keyPrefix != "" {
		buff.WriteString(r.keyPrefix)
	}
	buff.WriteString(query.Namespace())
	buff.WriteByte(':')
	buff.WriteString(query.Key())
	if query.Args() != "" {
		buff.WriteByte(':')
		buff.WriteString(query.Args())
	}
	return buff.String()
}

/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/20
   Description :
-------------------------------------------------
*/

package redis_cache

import (
	"context"
	"strconv"
)

// k-v. 批量删除匹配的key, 返回游标, 参数顺序: matchKey, cursor
const luaScanDel = `
-- 打印开始删除日志
if (KEYS[2] == "0") then redis.log(redis.LOG_NOTICE, "start scan del key:", KEYS[1]) end

-- 扫描key. 经过测试, count 取值 10000 不会导致长时间阻塞
local v = redis.call("SCAN", KEYS[2], "MATCH", KEYS[1], "COUNT", 10000)

-- 遍历key, 每100条进行一次删除
local ks, size = {}, 0
for key, value in ipairs(v[2]) do
    ks[size + 1], size = value, size + 1
    if (size == 100) then
        redis.call("DEL", unpack(ks))
        ks, size = {}, 0
    end
end

-- 删除剩下的key
if (size > 0) then
    redis.call("DEL", unpack(ks))
end

-- 打印结束删除日志
if (v[1] == "0") then redis.log(redis.LOG_NOTICE, "end scan del key:", KEYS[1]) end

-- 返回游标, 如果返回0 应该结束循环
return v[1]`

func (r *redisCache) scanDelKey(ctx context.Context, matchKey string) (err error) {
	var cursor int
	for {
		cursor, err = r.client.Eval(ctx, luaScanDel, []string{matchKey, strconv.Itoa(cursor)}).Int()
		if err != nil {
			return err
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}

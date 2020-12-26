/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package single_sf

import (
	"errors"
	"sync"

	"github.com/zlyuancn/zcache/core"
)

const (
	ShardCount uint64 = 1 << 5 // 分片数
)

type waitResult struct {
	wg sync.WaitGroup
	v  []byte
	e  error
}

var _ core.ISingleFlight = (*SingleFlight)(nil)

type SingleFlight struct {
	mxs        []*sync.RWMutex
	waits      []map[uint64]*waitResult
	shardCount uint64
	shardMod   uint64
}

// 创建一个单跑, 分片数必须大于0且为2的幂
func NewSingleFlight(shardCount ...uint64) *SingleFlight {
	count := ShardCount
	if len(shardCount) > 0 && shardCount[0] > 0 {
		count = shardCount[0]
		if count&(count-1) != 0 {
			panic(errors.New("shardCount must power of 2"))
		}
	}

	mxs := make([]*sync.RWMutex, count)
	mms := make([]map[uint64]*waitResult, count)
	for i := uint64(0); i < count; i++ {
		mxs[i] = new(sync.RWMutex)
		mms[i] = make(map[uint64]*waitResult)
	}
	return &SingleFlight{
		mxs:        mxs,
		waits:      mms,
		shardCount: count,
		shardMod:   count - 1,
	}
}

func (m *SingleFlight) Do(query core.IQuery, fn func(query core.IQuery) ([]byte, error)) ([]byte, error) {
	shard := query.GlobalId() & m.shardMod
	mx := m.mxs[shard]
	wait := m.waits[shard]

	mx.RLock()
	result, ok := wait[query.GlobalId()]
	mx.RUnlock()

	// 来晚了, 等待结果
	if ok {
		result.wg.Wait()
		return result.v, result.e
	}

	mx.Lock()

	// 再检查一下, 因为在拿到锁之前可能被别的进程占了位置
	result, ok = wait[query.GlobalId()]
	if ok {
		mx.Unlock()
		result.wg.Wait()
		return result.v, result.e
	}

	// 占位置
	result = new(waitResult)
	result.wg.Add(1)
	wait[query.GlobalId()] = result
	mx.Unlock()

	// 执行db加载
	result.v, result.e = fn(query)
	result.wg.Done()

	// 离开
	mx.Lock()
	delete(wait, query.GlobalId())
	mx.Unlock()

	return result.v, result.e
}

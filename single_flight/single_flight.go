/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/19
   Description :
-------------------------------------------------
*/

package single_flight

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
	mxs        []*sync.Mutex
	waits      []map[uint64]*waitResult
	shardCount uint64
	shardMod   uint64
}

// 创建一个单跑, 分片数必须大于0且为2的幂
func NewSingleFlight(shardCount ...uint64) *SingleFlight {
	count := ShardCount
	if len(shardCount) > 0 && shardCount[0] > 0 {
		count = shardCount[0]
		if count < 0 || count&(count-1) != 0 {
			panic(errors.New("shardCount must > 0 and shardCount must power of 2"))
		}
	}

	mxs := make([]*sync.Mutex, count)
	mms := make([]map[uint64]*waitResult, count)
	for i := uint64(0); i < count; i++ {
		mxs[i] = new(sync.Mutex)
		mms[i] = make(map[uint64]*waitResult)
	}
	return &SingleFlight{
		mxs:        mxs,
		waits:      mms,
		shardCount: count,
		shardMod:   count - 1,
	}
}

func (m *SingleFlight) Do(globalId uint64, fn func() ([]byte, error)) ([]byte, error) {
	shard := globalId & m.shardMod
	mx := m.mxs[shard]
	wait := m.waits[shard]

	mx.Lock()

	// 来晚了, 等待结果
	if c, ok := wait[globalId]; ok {
		mx.Unlock()
		c.wg.Wait()
		return c.v, c.e
	}

	// 占位置
	result := new(waitResult)
	result.wg.Add(1)
	wait[globalId] = result
	mx.Unlock()

	// 执行db加载
	result.v, result.e = fn()
	result.wg.Done()

	// 离开
	mx.Lock()
	delete(wait, globalId)
	mx.Unlock()

	return result.v, result.e
}

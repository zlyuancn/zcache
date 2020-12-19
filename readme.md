# 朴实无华的缓存模块

---

# 获得
`go get -u github.com/zlyuancn/zcache`

# db数据库
+ 支持任何数据库, 本模块不关心用户如何加载数据

# 缓存数据库
+ [任何实现 `cachedb.ICacheDB` 的结构](./core/cachedb.go)
+ [no-cache](./cachedb/no-cache/no-cache.go)
+ [memory-cache](./cachedb/memory-cache/memory-cache.go)
+ [redis](./cachedb/redis-cache/redis-cache.go)

# 编解码器

> 开发过程中不需要考虑每个对象的编解码, 可以在初始化时选择一个编解码器, 默认是`MsgPack`

+ [任何实现 `codec.ICodec` 的结构](./core/codec.go)
+ Byte
+ Json
+ JsonIterator
+ MsgPack
+ ProtoBuffer

# 如何解决缓存击穿

+ 可以在加载器中启用SingleFlight, 当有多个进程同时获取一个相同的数据时, 只有一个进程会真的去加载函数读取数据, 其他的进程会等待该进程结束直接收到结果. 可以通过 `core.ISingleFlight` 接口实现分布式锁让多个实例同一时间只有一个进程加载同一个数据.

# 如何解决缓存雪崩

+ 为加载器设置随机的TTL, 可以有效减小缓存雪崩的风险.

# 如何解决缓存穿透

+ 可以提供一个占位符存入缓存, 设置一个较小的TTL
+ 在用户请求key的时候预判断它是否可能不存在, 比如判断id长度不等于32(uuid去掉横杠的长度)的请求直接返回错误

# benchmark

> 未模拟用户请求和db加载, 直接测试本模块本身的性能

```shell script
go test -v -run "^$" -bench "^Benchmark.+$" -cpu 8,20,50,200,500 .
```

# 10 000 个key, 每个key 512字节随机数据, 请求key顺序随机

```text
CPU: 4c8t 3.7GHz
# memory-cache
BenchmarkMemoryCache_10k-8        	 4167741	       287 ns/op
BenchmarkMemoryCache_10k-20        	 3985488	       308 ns/op
BenchmarkMemoryCache_10k-50        	 3968851	       308 ns/op
BenchmarkMemoryCache_10k-200       	 3960068	       320 ns/op
BenchmarkMemoryCache_10k-500       	 3601851	       339 ns/op
BenchmarkRedisCache_10k-8         	   53870	     21670 ns/op
BenchmarkRedisCache_10k-20         	   68161	     17536 ns/op
BenchmarkRedisCache_10k-50         	   75562	     16189 ns/op
BenchmarkRedisCache_10k-200        	   62276	     18630 ns/op
BenchmarkRedisCache_10k-500        	   41538	     28168 ns/op
```

# 示例代码

```go
// 创建缓存服务
cache := zcache.NewCache()

// 注册加载器, 只有相同的 space 和 key 才会在加载数据时使用这个加载器
cache.RegisterLoader("test", "key", zcache.NewLoader(func(query zcache.IQuery) (interface{}, error) {
    return "hello " + query.ArgsText(), nil
}))

// 创建查询条件
q := zcache.NewQuery("test", "key", zcache.WithQueryArgs([]string{"arg1", "arg2"}))

// 创建用户保存结果的变量
var a string

// 从缓存加载, 如果加载失败会调用加载器, 随后将结果放入缓存中, 最后将数据写入a
_ = cache.Get(q, &a) // 接收变量必须传入指针

fmt.Println(a)
```

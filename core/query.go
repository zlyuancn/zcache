/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package core

// 查询参数
type IQuery interface {
	// 桶名, 不能为空, 在缓存未命中时, 用于决定使用哪个加载器
	Bucket() string
	// 参数, 不同的参数指向不同的数据
	Args() interface{}
	// 参数文本, 用于在缓存未命中时, 将Args传入加载器以查询数据, 这个文本是由Args()生成的
	ArgsText() string
	// 元数据, 元数据用于在上下文中传递数据, 不会参与key计算
	Meta() interface{}

	// 全局唯一id, 用于定位一条数据, 这个id是根据 Bucket 和 ArgsText 计算的
	GlobalId() uint64

	// 查询加载器, 缓存未命中时优先使用这个加载器生成数据
	Loader() ILoader
}

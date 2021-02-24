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
	// 命名空间, 不能为空, 用于找到储存数据的桶
	Namespace() string
	// key, 不能为空, 在缓存未命中时, 用于决定使用哪个加载器
	Key() string
	// 参数
	Args() interface{}
	// 参数文本, 用于在缓存未命中时, 将Args传入加载器以查询数据, 这个文本是由Args()生成的
	ArgsText() string
	// 获取元数据
	Meta() interface{}

	// 全局唯一id, 用于定位一条数据
	GlobalId() uint64

	// 查询加载器, 无数据时优先使用这个加载器
	Loader() ILoader
}

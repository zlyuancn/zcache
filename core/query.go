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
	// 参数, 用于在缓存未命中时, 将Args传入加载器以查询数据
	Args() interface{}
	// 获取元数据
	Meta() interface{}

	// 全局唯一id, 用于定位一条数据
	GlobalId() uint64
	// 参数文本, 它应该在创建Query时由Args计算并保存, 目的是为了一次获取数据过程中避免重复计算
	ArgsText() string
}

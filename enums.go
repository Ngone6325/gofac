package gofac

type LifetimeScope int

const (
	Transient LifetimeScope = iota // 瞬时：每次获取新建实例
	Singleton                      // 单例：全局唯一，根容器缓存
	Scoped                         // 作用域：作用域内唯一，不同作用域隔离
)

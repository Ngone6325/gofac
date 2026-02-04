package gofac

import "errors"

// 框架核心错误定义（新增Scoped相关错误，原有错误保留）
var (
	ErrNotFunc                   = errors.New("注册的必须是构造函数（函数类型）")
	ErrNoReturn                  = errors.New("构造函数必须有且仅有一个返回值")
	ErrRegisterDuplicate         = errors.New("该服务类型已注册，禁止重复注册")
	ErrServiceNotRegistered      = errors.New("服务未注册，无法解析")
	ErrCreateInstanceFailed      = errors.New("创建服务实例失败")
	ErrNotConcreteType           = errors.New("构造函数返回值必须是具体类型（非接口）")
	ErrResolveCircularDependency = errors.New("解析时发现循环依赖")
	ErrInvalidInterfaceType      = errors.New("interfaceType必须是接口的空指针，如(*IInterface)(nil)")
	ErrInvalidOutPtr             = errors.New("out必须是非空的指针类型")
	ErrTypeConvertFailed         = errors.New("实例无法转换为目标类型")
	ErrScopedOnRootContainer     = errors.New("Scoped生命周期服务不能直接通过根容器获取，请通过作用域Scope调用") // 新增Scoped错误
	ErrTransientInstance         = errors.New("实例注册不支持Transient生命周期，请使用Singleton或Scoped")
	ErrNilInstance               = errors.New("注册的实例不能为nil")
)

package di

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceDef 服务定义：存储注册元信息、缓存参数类型和单例实例
type ServiceDef struct {
	implType   reflect.Type   // 服务实现类型（构造函数返回值或实例类型）
	scope      LifetimeScope  // 生命周期
	instance   reflect.Value  // 单例实例缓存或预注册实例
	ctor       reflect.Value  // 构造函数反射值（实例注册时为空）
	ctorType   reflect.Type   // 构造函数反射类型（实例注册时为空）
	once       sync.Once      // 单例实例初始化原子操作
	paramTypes []reflect.Type // 缓存构造函数参数类型（核心优化）
	paramOnce  sync.Once      // 保证参数类型仅解析一次（并发安全）
	isInstance bool           // 是否为实例注册（true时直接使用instance，不调用ctor）
}

// Container DI容器核心：管理所有服务，保证并发安全
type Container struct {
	services map[reflect.Type]*ServiceDef
	mu       sync.RWMutex
}

// Scope 同一个Scope内Scoped实例唯一，不同Scope相互隔离
type Scope struct {
	root       *Container                     // 关联根容器（共享注册元信息）
	scopedInst map[reflect.Type]reflect.Value // 本作用域 Scoped 实例缓存
	mu         sync.RWMutex                   // 作用域并发安全锁
}

// NewContainer 创建新的DI容器
func NewContainer() *Container {
	return &Container{
		services: make(map[reflect.Type]*ServiceDef),
	}
}

// Global 全局容器：供单服务架构直接使用，省去手动创建容器
var Global = NewContainer()

// Register 基础注册：按构造函数返回值类型注册，返回错误（需手动处理）
func (c *Container) Register(ctor any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.register(ctor, nil, scope)
}

// RegisterAs 接口注册：将实现类型注册为指定接口类型，返回错误（需手动处理）
func (c *Container) RegisterAs(ctor any, interfaceType any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.register(ctor, interfaceType, scope)
}

// register 内部通用注册逻辑，抽离重复代码
func (c *Container) register(ctor any, interfaceType any, scope LifetimeScope) error {
	// 解析构造函数反射信息
	ctorVal := reflect.ValueOf(ctor)
	ctorType := ctorVal.Type()
	if ctorType.Kind() != reflect.Func {
		return ErrNotFunc
	}

	// 校验构造函数返回值：仅1个返回值，且为具体类型
	numOut := ctorType.NumOut()
	if numOut != 1 {
		return fmt.Errorf("%w，当前返回值数量：%d", ErrNoReturn, numOut)
	}
	implType := ctorType.Out(0)
	if implType.Kind() == reflect.Interface {
		return fmt.Errorf("%w，返回值为接口：%s", ErrNotConcreteType, implType)
	}

	// 确定最终注册的服务类型（接口/实现类型）
	svcType := implType
	if interfaceType != nil {
		// ---------- 修复核心：重写接口类型解析，增加空值校验，更健壮 ----------
		// 第一步：先通过Type获取接口类型，避免ValueOf(nil)的解析问题
		ifaceType := reflect.TypeOf(interfaceType)
		// 校验：必须是指针类型，且指向的元素是接口
		if ifaceType.Kind() != reflect.Ptr || ifaceType.Elem().Kind() != reflect.Interface {
			return ErrInvalidInterfaceType
		}
		// 提取真实的接口类型（指针指向的元素）
		svcType = ifaceType.Elem()
		// ---------------------------------------------------------------------

		// 校验实现类型是否实现接口
		if !implType.Implements(svcType) {
			return fmt.Errorf("类型%s未实现接口%s", implType, svcType)
		}
	}

	// 检查重复注册
	if _, exists := c.services[svcType]; exists {
		return fmt.Errorf("%w，类型：%s", ErrRegisterDuplicate, svcType)
	}

	// 封装服务定义并加入容器
	c.services[svcType] = &ServiceDef{
		implType:   implType,
		scope:      scope,
		ctor:       ctorVal,
		ctorType:   ctorType,
		isInstance: false,
	}
	return nil
}

// RegisterInstance 实例注册：直接注册已创建的实例，按实例类型注册
// 注意：不支持Transient生命周期（实例已创建，无法每次返回新实例）
func (c *Container) RegisterInstance(instance any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstance(instance, nil, scope)
}

// RegisterInstanceAs 实例接口注册：将已创建的实例注册为指定接口类型
// 注意：不支持Transient生命周期（实例已创建，无法每次返回新实例）
func (c *Container) RegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstance(instance, interfaceType, scope)
}

// registerInstance 内部实例注册逻辑
func (c *Container) registerInstance(instance any, interfaceType any, scope LifetimeScope) error {
	// Transient不支持实例注册（无法每次创建新实例）
	if scope == Transient {
		return ErrTransientInstance
	}

	// 校验实例不为 nil
	if instance == nil {
		return ErrNilInstance
	}

	instVal := reflect.ValueOf(instance)
	implType := instVal.Type()

	// 确定最终注册的服务类型（接口/实现类型）
	svcType := implType
	if interfaceType != nil {
		// 解析接口类型
		ifaceType := reflect.TypeOf(interfaceType)
		if ifaceType.Kind() != reflect.Ptr || ifaceType.Elem().Kind() != reflect.Interface {
			return ErrInvalidInterfaceType
		}
		svcType = ifaceType.Elem()

		// 校验实现类型是否实现接口
		if !implType.Implements(svcType) {
			return fmt.Errorf("实例类型%s未实现接口%s", implType, svcType)
		}
	}

	// 检查重复注册
	if _, exists := c.services[svcType]; exists {
		return fmt.Errorf("%w，类型：%s", ErrRegisterDuplicate, svcType)
	}

	// 封装服务定义并加入容器
	c.services[svcType] = &ServiceDef{
		implType:   implType,
		scope:      scope,
		instance:   instVal,
		isInstance: true,
	}
	return nil
}

// Resolve 原始解析：通过指针接收实例，返回错误（兼容旧逻辑）
func (c *Container) Resolve(out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}
	svcType := outVal.Elem().Type()
	instance, err := c.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return err
	}
	outVal.Elem().Set(instance)
	return nil
}

// resolve 内部递归解析核心方法：处理依赖、缓存、生命周期（原有逻辑新增Scoped校验）
func (c *Container) resolve(svcType reflect.Type, track map[reflect.Type]bool) (reflect.Value, error) {
	// 读锁获取服务定义，避免写阻塞
	c.mu.RLock()
	serviceDef, exists := c.services[svcType]
	c.mu.RUnlock()
	if !exists {
		return reflect.Value{}, fmt.Errorf("%w，类型：%s", ErrServiceNotRegistered, svcType)
	}

	// 循环依赖检测
	if track[svcType] {
		return reflect.Value{}, fmt.Errorf("%w，循环依赖链包含：%s", ErrResolveCircularDependency, svcType)
	}
	track[svcType] = true
	defer delete(track, svcType)

	// 新增：Scoped禁止根容器直接解析，强制使用作用域
	if serviceDef.scope == Scoped {
		return reflect.Value{}, ErrScopedOnRootContainer
	}

	// 实例注册：直接返回预注册的实例（Singleton/Scoped）
	if serviceDef.isInstance {
		return serviceDef.instance, nil
	}

	// 单例：已有实例直接返回
	if serviceDef.scope == Singleton && serviceDef.instance.IsValid() {
		return serviceDef.instance, nil
	}

	// 核心优化：缓存构造函数参数类型，仅首次解析
	serviceDef.paramOnce.Do(func() {
		numIn := serviceDef.ctorType.NumIn()
		params := make([]reflect.Type, numIn)
		for i := 0; i < numIn; i++ {
			params[i] = serviceDef.ctorType.In(i)
		}
		serviceDef.paramTypes = params
	})
	paramTypes := serviceDef.paramTypes

	// 递归解析所有依赖参数
	params := make([]reflect.Value, len(paramTypes))
	for i, pType := range paramTypes {
		pInstance, err := c.resolve(pType, track)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("解析依赖%s失败：%w", pType, err)
		}
		params[i] = pInstance
	}

	// 调用构造函数创建实例
	results := serviceDef.ctor.Call(params)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("%w，构造函数调用返回值异常", ErrCreateInstanceFailed)
	}
	instance := results[0]

	// 单例：原子操作缓存实例，保证仅创建一次
	if serviceDef.scope == Singleton {
		serviceDef.once.Do(func() {
			serviceDef.instance = instance
		})
	}

	return instance, nil
}

// NewScope 新增：Container创建作用域方法（根容器专属，创建Scoped作用域）
func (c *Container) NewScope() *Scope {
	return &Scope{
		root:       c,
		scopedInst: make(map[reflect.Type]reflect.Value),
	}
}

// Resolve 新增：Scope的Resolve方法（与Container的Resolve格式一致，支持Scoped）
func (s *Scope) Resolve(out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}
	svcType := outVal.Elem().Type()
	instance, err := s.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return err
	}
	outVal.Elem().Set(instance)
	return nil
}

// 新增：Scope的内部解析方法（处理所有生命周期，核心Scoped缓存逻辑）
func (s *Scope) resolve(svcType reflect.Type, track map[reflect.Type]bool) (reflect.Value, error) {
	// 从根容器获取注册元信息（所有作用域共享）
	s.root.mu.RLock()
	serviceDef, exists := s.root.services[svcType]
	s.root.mu.RUnlock()
	if !exists {
		return reflect.Value{}, fmt.Errorf("%w，类型：%s", ErrServiceNotRegistered, svcType)
	}

	// 循环依赖检测
	if track[svcType] {
		return reflect.Value{}, fmt.Errorf("%w，循环依赖链包含：%s", ErrResolveCircularDependency, svcType)
	}
	track[svcType] = true
	defer delete(track, svcType)

	// 实例注册处理
	if serviceDef.isInstance {
		// Singleton实例：直接返回根容器的实例
		if serviceDef.scope == Singleton {
			return serviceDef.instance, nil
		}
		// Scoped实例：每个作用域独立缓存
		if serviceDef.scope == Scoped {
			s.mu.RLock()
			inst, exists := s.scopedInst[svcType]
			s.mu.RUnlock()
			if exists && inst.IsValid() {
				return inst, nil
			}
			// 首次访问：缓存实例到作用域
			s.mu.Lock()
			s.scopedInst[svcType] = serviceDef.instance
			s.mu.Unlock()
			return serviceDef.instance, nil
		}
	}

	// 1. 单例：修复循环依赖 → 优先从根容器取缓存，未初始化则用作用域自身resolve解析（复用track）
	if serviceDef.scope == Singleton {
		// 读锁获取根容器的单例实例，已缓存则直接返回（核心：跳过根容器resolve，避免track重复写入）
		s.root.mu.RLock()
		if serviceDef.instance.IsValid() {
			inst := serviceDef.instance
			s.root.mu.RUnlock()
			return inst, nil
		}
		s.root.mu.RUnlock()
		// 单例未初始化：用作用域自身resolve完成初始化（复用当前track，无循环依赖误判）
		goto createInstance
	}

	// 2. Scoped：作用域内唯一，先查本作用域缓存
	if serviceDef.scope == Scoped {
		s.mu.RLock()
		inst, exists := s.scopedInst[svcType]
		s.mu.RUnlock()
		if exists && inst.IsValid() {
			return inst, nil
		}
	}

	// 新增标签：统一创建实例（Scoped/Transient/未初始化的Singleton共用）
createInstance:
	// 缓存未命中：解析参数+创建实例（Scoped/Transient/未初始化Singleton通用）
	serviceDef.paramOnce.Do(func() {
		numIn := serviceDef.ctorType.NumIn()
		params := make([]reflect.Type, numIn)
		for i := 0; i < numIn; i++ {
			params[i] = serviceDef.ctorType.In(i)
		}
		serviceDef.paramTypes = params
	})
	paramTypes := serviceDef.paramTypes

	params := make([]reflect.Value, len(paramTypes))
	for i, pType := range paramTypes {
		// 依赖解析：Scoped作用域内的依赖，优先通过作用域解析
		pInstance, err := s.resolve(pType, track)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("解析依赖%s失败：%w", pType, err)
		}
		params[i] = pInstance
	}

	results := serviceDef.ctor.Call(params)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("%w，构造函数调用返回值异常", ErrCreateInstanceFailed)
	}
	instance := results[0]

	// 3. Scoped：将实例写入本作用域缓存
	if serviceDef.scope == Scoped {
		s.mu.Lock()
		s.scopedInst[svcType] = instance
		s.mu.Unlock()
	}

	// 新增：未初始化的Singleton，创建后写入根容器缓存（保证全局唯一）
	if serviceDef.scope == Singleton {
		serviceDef.once.Do(func() {
			s.root.mu.Lock()
			serviceDef.instance = instance
			s.root.mu.Unlock()
		})
	}

	// 4. Transient：直接返回，不缓存
	return instance, nil
}

// getTyped 内部泛型解析：将反射获取的实例转换为目标类型T
func getTyped[T any](_ *Container, svcType reflect.Type, instance reflect.Value) (T, error) {
	var zero T
	// 处理接口类型、可赋值以及可转换类型
	it := instance.Type()
	// 如果目标类型是接口，检查实现关系
	if svcType.Kind() == reflect.Interface {
		// 情况1：实例类型直接实现接口（包括指针类型）
		if it.Implements(svcType) {
			return instance.Interface().(T), nil
		}
		// 情况2：值类型实现接口，但容器返回的是值 → 尝试取地址
		if it.Kind() != reflect.Ptr && reflect.PointerTo(it).Implements(svcType) {
			var iface any
			if instance.CanAddr() {
				iface = instance.Addr().Interface()
			} else {
				// 创建一个新的指针并设置值以便转换
				ptr := reflect.New(it)
				ptr.Elem().Set(instance)
				iface = ptr.Interface()
			}
			return iface.(T), nil
		}
		return zero, fmt.Errorf("【%w】实例%s无法转换为目标接口类型%s", ErrTypeConvertFailed, it, svcType)
	}

	// 目标不是接口：检查是否可直接赋值或可转换
	if it.AssignableTo(svcType) {
		return instance.Interface().(T), nil
	}
	if it.ConvertibleTo(svcType) {
		conv := instance.Convert(svcType)
		return conv.Interface().(T), nil
	}

	return zero, fmt.Errorf("【%w】实例%s无法转换为目标类型%s", ErrTypeConvertFailed, it, svcType)
}

// MustRegister ---------------------- 便捷Must系列方法（出错Panic，90%场景首选） ----------------------
// MustRegister 便捷基础注册：出错直接Panic
func (c *Container) MustRegister(ctor any, scope LifetimeScope) {
	if err := c.Register(ctor, scope); err != nil {
		panic(fmt.Sprintf("【DI注册失败】%v", err))
	}
}

// MustRegisterAs 便捷接口注册：出错直接Panic
func (c *Container) MustRegisterAs(ctor any, interfaceType any, scope LifetimeScope) {
	if err := c.RegisterAs(ctor, interfaceType, scope); err != nil {
		panic(fmt.Sprintf("【DI接口注册失败】%v", err))
	}
}

// MustRegisterInstance 便捷实例注册：出错直接Panic
func (c *Container) MustRegisterInstance(instance any, scope LifetimeScope) {
	if err := c.RegisterInstance(instance, scope); err != nil {
		panic(fmt.Sprintf("【DI实例注册失败】%v", err))
	}
}

// MustRegisterInstanceAs 便捷实例接口注册：出错直接Panic
func (c *Container) MustRegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) {
	if err := c.RegisterInstanceAs(instance, interfaceType, scope); err != nil {
		panic(fmt.Sprintf("【DI实例接口注册失败】%v", err))
	}
}

// MustResolve 便捷原始解析：出错直接Panic
func (c *Container) MustResolve(out any) {
	if err := c.Resolve(out); err != nil {
		panic(fmt.Sprintf("【DI解析失败】%v", err))
	}
}

// MustResolve 新增：Scope的MustResolve方法（与Container格式一致）
func (s *Scope) MustResolve(out any) {
	if err := s.Resolve(out); err != nil {
		panic(fmt.Sprintf("【DI作用域解析失败】%v", err))
	}
}

// MustRegister ---------------------- 全局容器顶层泛型函数（直接调用di.Get[T]()、di.MustGet[T]()，极致简洁） ----------------------
func MustRegister(ctor any, scope LifetimeScope) { Global.MustRegister(ctor, scope) }
func MustRegisterAs(ctor any, iface any, scope LifetimeScope) {
	Global.MustRegisterAs(ctor, iface, scope)
}
func MustRegisterInstance(instance any, scope LifetimeScope) {
	Global.MustRegisterInstance(instance, scope)
}
func MustRegisterInstanceAs(instance any, iface any, scope LifetimeScope) {
	Global.MustRegisterInstanceAs(instance, iface, scope)
}
func MustResolve(out any) { Global.MustResolve(out) }

// Get 泛型解析：直接返回实例，带错误处理（符合Go习惯）
func Get[T any]() (T, error) {
	var zero T
	svcType := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := Global.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return zero, fmt.Errorf("【DI获取失败】%w", err)
	}
	return getTyped[T](Global, svcType, instance)
}

// MustGet 泛型便捷解析：直接返回实例，出错Panic（推荐使用）
func MustGet[T any]() T {
	inst, err := Get[T]()
	if err != nil {
		panic(err)
	}
	return inst
}

// GlobalNewScope 新增：全局创建作用域的便捷方法
func GlobalNewScope() *Scope {
	return Global.NewScope()
}

// ScopeGet 新增：作用域版泛型Get - 传入Scope指针，实现Scoped生命周期泛型解析
func ScopeGet[T any](s *Scope) (T, error) {
	var zero T
	svcType := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := s.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return zero, fmt.Errorf("【DI作用域获取失败】%w", err)
	}
	return getTyped[T](s.root, svcType, instance)
}

// ScopeMustGet 新增：作用域版泛型MustGet - 传入Scope指针，出错Panic（推荐使用）
func ScopeMustGet[T any](s *Scope) T {
	inst, err := ScopeGet[T](s)
	if err != nil {
		panic(err)
	}
	return inst
}

// Reset 重置容器：清空所有服务和缓存（测试用）
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[reflect.Type]*ServiceDef)
}

// Reset 替换为👇 修复后代码
func (s *Scope) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock() // 正确：使用作用域自身的锁
	s.scopedInst = make(map[reflect.Type]reflect.Value)
}

// GlobalReset 重置全局容器（测试用）
func GlobalReset() { Global.Reset() }

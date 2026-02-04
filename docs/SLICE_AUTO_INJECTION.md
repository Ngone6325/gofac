# 切片自动注入功能文档

## 概述

gofac 容器现在支持**切片类型的自动注入**。当构造函数需要切片类型的参数时，容器会智能地处理依赖注入：

1. **优先使用已注册的切片实例**：如果切片类型本身已注册，直接使用
2. **自动收集元素类型实例**：如果切片类型未注册，自动收集所有该元素类型的实例（包括默认注册和命名注册）

## 使用场景

### 场景 1：直接注册切片（传统方式）

```go
type UserService struct {
    AllowedRoles []string
}

func NewUserService(roles []string) *UserService {
    return &UserService{AllowedRoles: roles}
}

func main() {
    container := di.NewContainer()

    // 直接注册切片实例
    roles := []string{"admin", "user", "guest"}
    container.MustRegisterInstance(roles, di.Singleton)

    // 注册依赖切片的服务
    container.MustRegister(NewUserService, di.Singleton)

    // 解析服务 - 会注入已注册的切片
    var service *UserService
    container.MustResolve(&service)

    fmt.Println(service.AllowedRoles) // [admin user guest]
}
```

### 场景 2：自动收集多个实例（新功能）

```go
type Database struct {
    Host string
}

type DatabaseManager struct {
    Databases []*Database
}

func NewDatabaseManager(dbs []*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}

func main() {
    container := di.NewContainer()

    // 注册多个数据库实例（使用命名注册）
    primary := &Database{Host: "primary.db"}
    replica1 := &Database{Host: "replica1.db"}
    replica2 := &Database{Host: "replica2.db"}

    container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
    container.MustRegisterInstanceNamed("replica1", replica1, di.Singleton)
    container.MustRegisterInstanceNamed("replica2", replica2, di.Singleton)

    // 注册 DatabaseManager
    // 容器会自动收集所有 *Database 实例并作为切片注入
    container.MustRegister(NewDatabaseManager, di.Singleton)

    // 解析服务
    var manager *DatabaseManager
    container.MustResolve(&manager)

    fmt.Printf("Total databases: %d\n", len(manager.Databases)) // 3
}
```

## 工作原理

### 解析逻辑

当容器解析构造函数的依赖参数时：

```
1. 检查参数类型是否为切片
   ├─ 是切片类型
   │  ├─ 检查切片类型本身是否已注册（如 []string）
   │  │  ├─ 已注册 → 直接解析并注入该切片实例
   │  │  └─ 未注册 → 自动收集元素类型的所有实例
   │  │     ├─ 收集默认注册的实例（如果存在）
   │  │     ├─ 收集所有命名注册的实例
   │  │     └─ 组合成切片并注入
   │  └─ 返回切片
   └─ 非切片类型 → 正常解析
```

### 示例流程

```go
// 场景：未注册 []*Database 切片类型

container.MustRegisterInstanceNamed("db1", &Database{Host: "db1"}, di.Singleton)
container.MustRegisterInstanceNamed("db2", &Database{Host: "db2"}, di.Singleton)

// 当解析 NewDatabaseManager([]*Database) 时：
// 1. 检测到参数类型是 []*Database
// 2. 检查 []*Database 是否已注册 → 未注册
// 3. 提取元素类型 *Database
// 4. 收集所有 *Database 实例：
//    - 检查默认注册（无）
//    - 检查命名注册（找到 db1, db2）
// 5. 组合成 []*Database{db1, db2}
// 6. 注入到构造函数
```

## 完整示例

### 示例 1：插件系统

```go
type IPlugin interface {
    Initialize()
    GetName() string
}

type Plugin1 struct {
    Name string
}

func (p *Plugin1) Initialize() {
    fmt.Printf("Initializing %s\n", p.Name)
}

func (p *Plugin1) GetName() string {
    return p.Name
}

type PluginManager struct {
    Plugins []IPlugin
}

func NewPluginManager(plugins []IPlugin) *PluginManager {
    return &PluginManager{Plugins: plugins}
}

func main() {
    container := di.NewContainer()

    // 注册多个插件
    plugin1 := &Plugin1{Name: "Plugin1"}
    plugin2 := &Plugin1{Name: "Plugin2"}
    plugin3 := &Plugin1{Name: "Plugin3"}

    container.MustRegisterInstanceAsNamed("plugin1", plugin1, (*IPlugin)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("plugin2", plugin2, (*IPlugin)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("plugin3", plugin3, (*IPlugin)(nil), di.Singleton)

    // 注册插件管理器 - 自动注入所有插件
    container.MustRegister(NewPluginManager, di.Singleton)

    // 解析并初始化所有插件
    var manager *PluginManager
    container.MustResolve(&manager)

    for _, plugin := range manager.Plugins {
        plugin.Initialize()
    }
}
```

**输出：**
```
Initializing Plugin1
Initializing Plugin2
Initializing Plugin3
```

### 示例 2：中间件链

```go
type IMiddleware interface {
    Handle(ctx *Context) error
}

type AuthMiddleware struct{}
type LoggingMiddleware struct{}
type CorsMiddleware struct{}

func (m *AuthMiddleware) Handle(ctx *Context) error {
    fmt.Println("Auth middleware")
    return nil
}

func (m *LoggingMiddleware) Handle(ctx *Context) error {
    fmt.Println("Logging middleware")
    return nil
}

func (m *CorsMiddleware) Handle(ctx *Context) error {
    fmt.Println("CORS middleware")
    return nil
}

type MiddlewareChain struct {
    Middlewares []IMiddleware
}

func NewMiddlewareChain(middlewares []IMiddleware) *MiddlewareChain {
    return &MiddlewareChain{Middlewares: middlewares}
}

func main() {
    container := di.NewContainer()

    // 注册多个中间件
    container.MustRegisterInstanceAsNamed("auth", &AuthMiddleware{}, (*IMiddleware)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("logging", &LoggingMiddleware{}, (*IMiddleware)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("cors", &CorsMiddleware{}, (*IMiddleware)(nil), di.Singleton)

    // 注册中间件链 - 自动注入所有中间件
    container.MustRegister(NewMiddlewareChain, di.Singleton)

    // 解析并执行中间件链
    var chain *MiddlewareChain
    container.MustResolve(&chain)

    ctx := &Context{}
    for _, middleware := range chain.Middlewares {
        middleware.Handle(ctx)
    }
}
```

## 优先级规则

1. **切片类型已注册** → 使用已注册的切片实例
2. **切片类型未注册** → 自动收集元素类型的实例

这个设计确保了：
- ✅ **向后兼容**：现有代码继续工作
- ✅ **灵活性**：可以选择直接注册切片或让容器自动收集
- ✅ **智能化**：容器自动处理复杂的依赖关系

## 注意事项

### 1. 空切片

如果没有注册任何元素类型的实例，会注入空切片：

```go
func NewService(items []string) *Service {
    return &Service{Items: items}
}

container.MustRegister(NewService, di.Singleton)

var service *Service
container.MustResolve(&service)

fmt.Println(len(service.Items)) // 0 - 空切片
```

### 2. 顺序不保证

自动收集的实例顺序不保证，不要依赖特定顺序：

```go
// ❌ 不要依赖顺序
manager.Databases[0] // 可能是任何一个数据库

// ✅ 通过属性识别
for _, db := range manager.Databases {
    if db.IsPrimary {
        // 使用主库
    }
}
```

### 3. 混合使用

可以同时使用默认注册和命名注册：

```go
// 默认注册
defaultDB := &Database{Host: "default"}
container.MustRegisterInstance(defaultDB, di.Singleton)

// 命名注册
primary := &Database{Host: "primary"}
container.MustRegisterInstanceNamed("primary", primary, di.Singleton)

// 自动收集时会包含两者
container.MustRegister(NewDatabaseManager, di.Singleton)
// manager.Databases 包含 defaultDB 和 primary
```

## 最佳实践

### 1. 明确意图

如果需要特定的切片内容，直接注册切片：

```go
// ✅ 明确的切片内容
roles := []string{"admin", "user"}
container.MustRegisterInstance(roles, di.Singleton)
```

如果需要收集所有实例，使用命名注册：

```go
// ✅ 收集所有数据库
container.MustRegisterInstanceNamed("db1", db1, di.Singleton)
container.MustRegisterInstanceNamed("db2", db2, di.Singleton)
// 构造函数会自动收集
```

### 2. 使用接口

对于插件系统，使用接口类型：

```go
type IPlugin interface {
    Initialize()
}

// 注册为接口类型
container.MustRegisterInstanceAsNamed("plugin1", plugin1, (*IPlugin)(nil), di.Singleton)

// 构造函数接收接口切片
func NewPluginManager(plugins []IPlugin) *PluginManager {
    return &PluginManager{Plugins: plugins}
}
```

### 3. 文档化行为

在构造函数注释中说明切片参数的来源：

```go
// NewDatabaseManager 创建数据库管理器
// dbs: 自动注入所有已注册的 *Database 实例（包括命名注册）
func NewDatabaseManager(dbs []*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}
```

## 与 ResolveAll 的区别

| 特性 | 切片自动注入 | ResolveAll |
|------|-------------|-----------|
| 使用场景 | 构造函数依赖 | 手动解析 |
| 调用方式 | 自动 | 手动调用 |
| 优先级 | 优先使用已注册切片 | 总是收集所有实例 |
| 适用范围 | 依赖注入 | 任何时候 |

```go
// 切片自动注入（自动）
container.MustRegister(NewDatabaseManager, di.Singleton)
var manager *DatabaseManager
container.MustResolve(&manager)

// ResolveAll（手动）
var dbs []*Database
container.MustResolveAll(&dbs)
```

## 完整示例代码

完整的示例代码请参考 `main.go` 中的示例 9：

```go
// 示例9：命名注册 + 嵌套依赖注入
func ExampleMoreComplexReferenceTypes() {
    container := di.NewContainer()

    // 使用命名注册来注册多个同类型的实例
    service1 := NewUserService([]string{"admin", "user", "guest"})
    service2 := NewUserService([]string{"admin2", "user2", "guest2"})
    service3 := NewUserService([]string{"admin3", "user3", "guest3"})

    container.MustRegisterInstanceNamed("service1", service1, di.Singleton)
    container.MustRegisterInstanceNamed("service2", service2, di.Singleton)
    container.MustRegisterInstanceNamed("service3", service3, di.Singleton)

    // 注册 ArrayUserService - 自动注入所有 *UserService 实例
    container.MustRegister(NewArrayUserService, di.Singleton)

    // 注册 ArrayUserService2 - 自动注入 *ArrayUserService
    container.MustRegister(NewArrayUserService2, di.Singleton)

    // 解析嵌套结构
    var arrayUserService2 *ArrayUserService2
    container.MustResolve(&arrayUserService2)

    fmt.Printf("Total services: %d\n", len(arrayUserService2.Services.Services))
}
```

---

**版本**: v1.4.0
**更新日期**: 2026-02-02

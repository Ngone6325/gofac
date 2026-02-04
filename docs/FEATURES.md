# Gofac 新增特性文档

本文档介绍 gofac 项目新增的 `RegisterInstance` 方法和引用类型支持。

## 目录

1. [RegisterInstance 实例注册](#registerinstance-实例注册)
2. [引用类型支持](#引用类型支持)
3. [API 参考](#api-参考)
4. [使用示例](#使用示例)

---

## RegisterInstance 实例注册

### 概述

`RegisterInstance` 系列方法允许你直接注册已创建的实例，而不是注册构造函数。这在以下场景非常有用：

- **预配置对象**：配置对象、连接池等已经初始化的对象
- **外部依赖**：从外部系统获取的对象
- **测试模拟**：在单元测试中注入 mock 对象
- **第三方库对象**：无法通过构造函数创建的对象

### 支持的生命周期

| 生命周期 | 支持 | 说明 |
|---------|------|------|
| `Singleton` | ✅ | 全局唯一实例，所有解析返回同一个实例 |
| `Scoped` | ✅ | 每个作用域共享同一个实例 |
| `Transient` | ❌ | 不支持（实例已创建，无法每次返回新实例） |

### 方法列表

#### 容器方法

```go
// 基础方法（返回错误）
func (c *Container) RegisterInstance(instance any, scope LifetimeScope) error
func (c *Container) RegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) error

// Must 方法（出错 panic）
func (c *Container) MustRegisterInstance(instance any, scope LifetimeScope)
func (c *Container) MustRegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope)
```

#### 全局容器方法

```go
func MustRegisterInstance(instance any, scope LifetimeScope)
func MustRegisterInstanceAs(instance any, iface any, scope LifetimeScope)
```

### 错误处理

新增错误类型：

- `ErrTransientInstance`：实例注册不支持 Transient 生命周期
- `ErrNilInstance`：注册的实例不能为 nil

---

## 引用类型支持

### 概述

gofac 现在完全支持 Go 的引用类型作为依赖注入的服务类型，包括：

- **切片（Slice）**：`[]T`
- **映射（Map）**：`map[K]V`
- **数组（Array）**：`[N]T`

这些类型可以：
- 作为服务类型注册
- 作为构造函数的依赖参数
- 通过 `RegisterInstance` 直接注册实例
- 通过构造函数返回值注册

### 支持的类型

| 类型 | 示例 | 支持 |
|------|------|------|
| 切片 | `[]string`, `[]int`, `[]*User` | ✅ |
| 映射 | `map[string]int`, `map[int]*Config` | ✅ |
| 数组 | `[5]int`, `[10]string` | ✅ |
| 结构体 | `User`, `*Config` | ✅ |
| 接口 | `ILogger`, `IRepository` | ✅ |
| 基础类型 | `int`, `string`, `bool` | ✅ |

### 注意事项

1. **值语义 vs 引用语义**
   - 切片、映射：引用类型，多个解析共享底层数据
   - 数组：值类型，会复制整个数组

2. **并发安全**
   - 容器本身是线程安全的
   - 但注册的切片/映射实例本身不是线程安全的
   - 如需并发访问，请使用 `sync.Map` 或加锁

3. **生命周期**
   - `Singleton`：全局共享同一个切片/映射实例
   - `Scoped`：每个作用域共享同一个实例
   - `Transient`：每次返回新实例（仅构造函数注册支持）

---

## API 参考

### RegisterInstance

注册已创建的实例，按实例类型注册。

```go
func (c *Container) RegisterInstance(instance any, scope LifetimeScope) error
```

**参数：**
- `instance`：要注册的实例（不能为 nil）
- `scope`：生命周期（`Singleton` 或 `Scoped`）

**返回：**
- `error`：注册失败时返回错误

**示例：**

```go
config := &Config{AppName: "MyApp", Port: 8080}
err := container.RegisterInstance(config, di.Singleton)
if err != nil {
    log.Fatal(err)
}
```

### RegisterInstanceAs

注册已创建的实例为指定接口类型。

```go
func (c *Container) RegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) error
```

**参数：**
- `instance`：要注册的实例（不能为 nil）
- `interfaceType`：接口类型，格式为 `(*IInterface)(nil)`
- `scope`：生命周期（`Singleton` 或 `Scoped`）

**返回：**
- `error`：注册失败时返回错误

**示例：**

```go
logger := &ConsoleLogger{Prefix: "INFO"}
err := container.RegisterInstanceAs(logger, (*ILogger)(nil), di.Singleton)
if err != nil {
    log.Fatal(err)
}
```

### MustRegisterInstance

便捷方法，注册失败时 panic。

```go
func (c *Container) MustRegisterInstance(instance any, scope LifetimeScope)
```

**示例：**

```go
config := &Config{AppName: "MyApp", Port: 8080}
container.MustRegisterInstance(config, di.Singleton)
```

### MustRegisterInstanceAs

便捷方法，注册失败时 panic。

```go
func (c *Container) MustRegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope)
```

**示例：**

```go
logger := &ConsoleLogger{Prefix: "INFO"}
container.MustRegisterInstanceAs(logger, (*ILogger)(nil), di.Singleton)
```

---

## 使用示例

### 示例 1：基础实例注册

```go
package main

import (
    "fmt"
    "gofac/di"
)

type Config struct {
    AppName string
    Port    int
}

func main() {
    container := di.NewContainer()

    // 创建配置实例
    config := &Config{
        AppName: "MyApp",
        Port:    8080,
    }

    // 注册实例为单例
    container.MustRegisterInstance(config, di.Singleton)

    // 解析获取实例
    var resolvedConfig *Config
    container.MustResolve(&resolvedConfig)

    fmt.Printf("AppName: %s, Port: %d\n", resolvedConfig.AppName, resolvedConfig.Port)
    fmt.Printf("Same instance: %v\n", config == resolvedConfig)
}
```

**输出：**
```
AppName: MyApp, Port: 8080
Same instance: true
```

### 示例 2：接口实例注册

```go
type ILogger interface {
    Log(msg string)
}

type ConsoleLogger struct {
    Prefix string
}

func (l *ConsoleLogger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.Prefix, msg)
}

func main() {
    container := di.NewContainer()

    // 创建日志实例
    logger := &ConsoleLogger{Prefix: "INFO"}

    // 注册实例为接口类型
    container.MustRegisterInstanceAs(logger, (*ILogger)(nil), di.Singleton)

    // 通过接口类型解析
    var resolvedLogger ILogger
    container.MustResolve(&resolvedLogger)

    resolvedLogger.Log("Hello from instance registration!")
}
```

**输出：**
```
[INFO] Hello from instance registration!
```

### 示例 3：切片类型支持

```go
type UserService struct {
    AllowedRoles []string
}

func NewUserService(roles []string) *UserService {
    return &UserService{AllowedRoles: roles}
}

func main() {
    container := di.NewContainer()

    // 注册切片实例
    roles := []string{"admin", "user", "guest"}
    container.MustRegisterInstance(roles, di.Singleton)

    // 注册依赖切片的服务
    container.MustRegister(NewUserService, di.Singleton)

    // 解析服务
    var userService *UserService
    container.MustResolve(&userService)

    fmt.Printf("Allowed roles: %v\n", userService.AllowedRoles)
}
```

**输出：**
```
Allowed roles: [admin user guest]
```

### 示例 4：Map 类型支持

```go
type ConfigService struct {
    Settings map[string]string
}

func NewConfigService(settings map[string]string) *ConfigService {
    return &ConfigService{Settings: settings}
}

func main() {
    container := di.NewContainer()

    // 注册 map 实例
    settings := map[string]string{
        "db_host": "localhost",
        "db_port": "5432",
        "db_name": "mydb",
    }
    container.MustRegisterInstance(settings, di.Singleton)

    // 注册依赖 map 的服务
    container.MustRegister(NewConfigService, di.Singleton)

    // 解析服务
    var configService *ConfigService
    container.MustResolve(&configService)

    fmt.Printf("DB Host: %s\n", configService.Settings["db_host"])
}
```

**输出：**
```
DB Host: localhost
```

### 示例 5：数组类型支持

```go
type PriorityQueue struct {
    Priorities [5]int
}

func NewPriorityQueue(priorities [5]int) *PriorityQueue {
    return &PriorityQueue{Priorities: priorities}
}

func main() {
    container := di.NewContainer()

    // 注册数组实例
    priorities := [5]int{1, 2, 3, 4, 5}
    container.MustRegisterInstance(priorities, di.Singleton)

    // 注册依赖数组的服务
    container.MustRegister(NewPriorityQueue, di.Singleton)

    // 解析服务
    var queue *PriorityQueue
    container.MustResolve(&queue)

    fmt.Printf("Priorities: %v\n", queue.Priorities)
}
```

**输出：**
```
Priorities: [1 2 3 4 5]
```

### 示例 6：Scoped 实例注册

```go
type RequestContext struct {
    RequestID string
}

func main() {
    container := di.NewContainer()

    // 注册为 Scoped（每个作用域独立）
    ctx := &RequestContext{RequestID: "req-001"}
    container.MustRegisterInstance(ctx, di.Scoped)

    // 创建两个作用域
    scope1 := container.NewScope()
    scope2 := container.NewScope()

    // 从 scope1 解析
    var ctx1 *RequestContext
    scope1.MustResolve(&ctx1)
    fmt.Printf("Scope1 RequestID: %s\n", ctx1.RequestID)

    // 从 scope2 解析
    var ctx2 *RequestContext
    scope2.MustResolve(&ctx2)
    fmt.Printf("Scope2 RequestID: %s\n", ctx2.RequestID)

    // 两个作用域获取的是同一个实例
    fmt.Printf("Same instance: %v\n", ctx1 == ctx2)
}
```

**输出：**
```
Scope1 RequestID: req-001
Scope2 RequestID: req-001
Same instance: true
```

### 示例 7：复杂引用类型组合

```go
type DatabaseConfig struct {
    Hosts    []string
    Ports    map[string]int
    Replicas [3]string
}

type Database struct {
    Config *DatabaseConfig
}

func NewDatabase(config *DatabaseConfig) *Database {
    return &Database{Config: config}
}

func main() {
    container := di.NewContainer()

    // 注册复杂配置实例
    dbConfig := &DatabaseConfig{
        Hosts:    []string{"host1", "host2", "host3"},
        Ports:    map[string]int{"primary": 5432, "replica": 5433},
        Replicas: [3]string{"replica1", "replica2", "replica3"},
    }
    container.MustRegisterInstance(dbConfig, di.Singleton)

    // 注册数据库服务
    container.MustRegister(NewDatabase, di.Singleton)

    // 解析服务
    var db *Database
    container.MustResolve(&db)

    fmt.Printf("Hosts: %v\n", db.Config.Hosts)
    fmt.Printf("Primary Port: %d\n", db.Config.Ports["primary"])
    fmt.Printf("Replicas: %v\n", db.Config.Replicas)
}
```

**输出：**
```
Hosts: [host1 host2 host3]
Primary Port: 5432
Replicas: [replica1 replica2 replica3]
```

### 示例 8：全局容器便捷方法

```go
func main() {
    // 使用全局容器注册实例
    config := &Config{AppName: "GlobalApp", Port: 9000}
    di.MustRegisterInstance(config, di.Singleton)

    // 使用泛型方法解析
    resolvedConfig := di.MustGet[*Config]()

    fmt.Printf("AppName: %s, Port: %d\n", resolvedConfig.AppName, resolvedConfig.Port)
}
```

**输出：**
```
AppName: GlobalApp, Port: 9000
```

---

## 最佳实践

### 1. 何时使用 RegisterInstance

✅ **推荐使用场景：**
- 配置对象（从文件/环境变量加载）
- 数据库连接池（已初始化）
- 第三方库对象（无构造函数）
- 测试中的 mock 对象

❌ **不推荐使用场景：**
- 需要延迟初始化的对象（使用构造函数注册）
- 需要每次创建新实例的对象（使用 Transient + 构造函数）
- 有复杂依赖关系的对象（使用构造函数注册，让容器管理依赖）

### 2. 引用类型的并发安全

```go
// ❌ 不安全：多个 goroutine 同时修改
settings := map[string]string{"key": "value"}
container.MustRegisterInstance(settings, di.Singleton)

// ✅ 安全：使用 sync.Map
var settings sync.Map
container.MustRegisterInstance(&settings, di.Singleton)

// ✅ 安全：只读访问
roles := []string{"admin", "user"}  // 注册后不修改
container.MustRegisterInstance(roles, di.Singleton)
```

### 3. 生命周期选择

| 场景 | 推荐生命周期 |
|------|-------------|
| 全局配置 | `Singleton` |
| 数据库连接池 | `Singleton` |
| HTTP 请求上下文 | `Scoped` |
| 用户会话数据 | `Scoped` |
| 临时数据 | 不推荐使用实例注册 |

---

## 完整 API 列表

### 注册方法

| 方法 | 说明 | 返回错误 |
|------|------|---------|
| `Register` | 构造函数注册 | ✅ |
| `RegisterAs` | 构造函数接口注册 | ✅ |
| `RegisterInstance` | 实例注册 | ✅ |
| `RegisterInstanceAs` | 实例接口注册 | ✅ |
| `MustRegister` | 构造函数注册（panic） | ❌ |
| `MustRegisterAs` | 构造函数接口注册（panic） | ❌ |
| `MustRegisterInstance` | 实例注册（panic） | ❌ |
| `MustRegisterInstanceAs` | 实例接口注册（panic） | ❌ |

### 解析方法

| 方法 | 说明 | 返回错误 |
|------|------|---------|
| `Resolve` | 指针解析 | ✅ |
| `MustResolve` | 指针解析（panic） | ❌ |
| `Get[T]` | 泛型解析 | ✅ |
| `MustGet[T]` | 泛型解析（panic） | ❌ |
| `ScopeGet[T]` | 作用域泛型解析 | ✅ |
| `ScopeMustGet[T]` | 作用域泛型解析（panic） | ❌ |

---

## 更新日志

### v1.1.0 (2026-02-02)

**新增功能：**
- ✨ 新增 `RegisterInstance` 和 `RegisterInstanceAs` 方法
- ✨ 新增 `MustRegisterInstance` 和 `MustRegisterInstanceAs` 便捷方法
- ✨ 完整支持切片、映射、数组等引用类型
- ✨ 实例注册支持 `Singleton` 和 `Scoped` 生命周期

**错误处理：**
- 🛡️ 新增 `ErrTransientInstance` 错误
- 🛡️ 新增 `ErrNilInstance` 错误

**文档：**
- 📚 新增 `FEATURES.md` 完整特性文档
- 📚 新增 `example_demo.go` 示例代码

---

## 常见问题

### Q1: 为什么实例注册不支持 Transient？

**A:** Transient 生命周期要求每次解析都返回新实例，但实例注册时对象已经创建，无法每次创建新实例。如果需要 Transient 行为，请使用构造函数注册。

### Q2: 切片/映射是引用类型，会有并发问题吗？

**A:** 容器本身是线程安全的，但注册的切片/映射实例本身不是线程安全的。如果多个 goroutine 会修改这些数据，请使用 `sync.Map` 或加锁保护。

### Q3: Scoped 实例注册和构造函数注册有什么区别？

**A:**
- **实例注册**：所有作用域共享同一个预创建的实例
- **构造函数注册**：每个作用域调用构造函数创建独立的实例

### Q4: 可以注册 nil 实例吗？

**A:** 不可以。注册 nil 实例会返回 `ErrNilInstance` 错误。

### Q5: 引用类型支持泛型解析吗？

**A:** 完全支持！可以使用 `Get[[]string]()` 或 `MustGet[map[string]int]()` 等泛型方法。

---

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

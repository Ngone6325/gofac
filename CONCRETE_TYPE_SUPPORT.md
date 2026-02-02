# RegisterAs 和 RegisterInstanceAs 支持具体类型

## 概述

`RegisterAs` 和 `RegisterInstanceAs` 方法现在同时支持**接口类型**和**具体类型**作为目标类型。

## 使用方式

### 1. 注册为接口类型（原有功能）

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

func NewConsoleLogger() *ConsoleLogger {
    return &ConsoleLogger{Prefix: "INFO"}
}

// 构造函数注册为接口
container.MustRegisterAs(NewConsoleLogger, (*ILogger)(nil), di.Singleton)

// 实例注册为接口
logger := &ConsoleLogger{Prefix: "DEBUG"}
container.MustRegisterInstanceAs(logger, (*ILogger)(nil), di.Singleton)

// 通过接口类型解析
var resolved ILogger
container.MustResolve(&resolved)
```

### 2. 注册为具体类型（新功能）⭐

```go
type UserService struct {
    Name string
}

func NewUserService() *UserService {
    return &UserService{Name: "service"}
}

// 构造函数注册为具体类型
container.MustRegisterAs(NewUserService, (*UserService)(nil), di.Singleton)

// 实例注册为具体类型
service := &UserService{Name: "instance"}
container.MustRegisterInstanceAs(service, (*UserService)(nil), di.Singleton)

// 通过具体类型解析
var resolved *UserService
container.MustResolve(&resolved)
```

## 使用场景

### 场景 1：同一个实现注册为多个类型

```go
type MixedService struct {
    Value string
}

func (m *MixedService) GetValue() string {
    return m.Value
}

type IService interface {
    GetValue() string
}

// 注册为接口类型
container.MustRegisterAs(NewMixedService, (*IService)(nil), di.Singleton)

// 同时注册为具体类型（使用不同的构造函数）
newConcrete := func() *MixedService {
    return &MixedService{Value: "concrete"}
}
container.MustRegisterAs(newConcrete, (*MixedService)(nil), di.Singleton)

// 可以通过两种方式解析
var iface IService
container.MustResolve(&iface)

var concrete *MixedService
container.MustResolve(&concrete)
```

### 场景 2：复杂依赖图中的类型控制

```go
type ServiceA struct {
    Name string
}

type ServiceB struct {
    A *ServiceA
}

type ServiceC struct {
    A *ServiceA
    B *ServiceB
}

// 全部注册为具体类型，精确控制依赖关系
container.MustRegisterAs(NewServiceA, (*ServiceA)(nil), di.Singleton)
container.MustRegisterAs(NewServiceB, (*ServiceB)(nil), di.Singleton)
container.MustRegisterAs(NewServiceC, (*ServiceC)(nil), di.Singleton)

// ServiceC 的依赖会自动解析
var c *ServiceC
container.MustResolve(&c)
```

### 场景 3：测试中的类型替换

```go
// 生产代码
type RealService struct {
    DB *Database
}

// 测试代码
type MockService struct {
    Data map[string]string
}

// 在测试中，可以注册 MockService 为 RealService 类型
// 这样依赖 *RealService 的代码无需修改
mockService := &MockService{Data: map[string]string{"key": "value"}}
container.MustRegisterInstanceAs(mockService, (*RealService)(nil), di.Singleton)
```

## 类型兼容性

注册为具体类型时，系统会检查类型兼容性：

- ✅ **相同类型**：`*UserService` -> `*UserService`
- ✅ **可赋值类型**：支持类型转换
- ✅ **指针/值类型转换**：自动处理指针和值类型的转换
- ❌ **不兼容类型**：会返回错误

## 语法说明

### 接口类型语法

```go
// 使用 (*InterfaceName)(nil) 表示接口类型
container.RegisterAs(ctor, (*ILogger)(nil), scope)
```

- `(*ILogger)(nil)` 是一个指向接口的空指针
- 系统会提取 `ILogger` 接口类型作为服务类型

### 具体类型语法

```go
// 使用 (*StructName)(nil) 表示具体类型
container.RegisterAs(ctor, (*UserService)(nil), scope)
```

- `(*UserService)(nil)` 是一个指向结构体的空指针
- 系统会使用 `*UserService` 指针类型作为服务类型

## 完整示例

```go
package main

import (
    "fmt"
    "gofac/di"
)

// 定义服务
type UserService struct {
    Name string
}

func NewUserService() *UserService {
    return &UserService{Name: "user-service"}
}

// 定义依赖 UserService 的服务
type OrderService struct {
    UserService *UserService
}

func NewOrderService(us *UserService) *OrderService {
    return &OrderService{UserService: us}
}

func main() {
    container := di.NewContainer()

    // 注册 UserService 为具体类型
    container.MustRegisterAs(NewUserService, (*UserService)(nil), di.Singleton)

    // 注册 OrderService（会自动解析 *UserService 依赖）
    container.MustRegister(NewOrderService, di.Singleton)

    // 解析 OrderService
    var orderService *OrderService
    container.MustResolve(&orderService)

    fmt.Printf("User: %s\n", orderService.UserService.Name)
    // 输出: User: user-service
}
```

## 与 Register 方法的区别

| 方法 | 注册类型 | 说明 |
|------|---------|------|
| `Register(ctor, scope)` | 自动推断 | 使用构造函数返回值类型 |
| `RegisterAs(ctor, (*Interface)(nil), scope)` | 接口类型 | 注册为指定接口类型 |
| `RegisterAs(ctor, (*Struct)(nil), scope)` | 具体类型 | 注册为指定具体类型 |

## 注意事项

1. **类型必须兼容**：注册的实现类型必须与目标类型兼容
2. **避免重复注册**：同一类型只能注册一次
3. **指针类型**：使用 `(*Type)(nil)` 语法时，会注册为指针类型 `*Type`
4. **依赖解析**：依赖注入时会按照注册的类型进行匹配

## 错误处理

```go
// 类型不兼容会返回错误
type ServiceA struct{}
type ServiceB struct{}

err := container.RegisterAs(
    func() *ServiceA { return &ServiceA{} },
    (*ServiceB)(nil),
    di.Singleton,
)
// err: 类型*ServiceA无法转换为目标类型*ServiceB
```

## 测试

完整的测试用例请参考 `di/concrete_type_test.go`，包括：

- ✅ 构造函数注册为具体类型
- ✅ 实例注册为具体类型
- ✅ 具体类型作为依赖注入
- ✅ 同时注册接口和具体类型
- ✅ 复杂依赖图
- ✅ 泛型方法解析

---

**版本**: v1.2.0
**更新日期**: 2026-02-02

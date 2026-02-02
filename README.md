# Gofac - Go 依赖注入容器

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Gofac 是一个受 [Autofac](https://autofac.org/) 启发的 Go 语言依赖注入（DI）容器，提供简洁、类型安全的依赖管理方案。

## ✨ 特性

- 🚀 **三种生命周期**：Transient（瞬时）、Singleton（单例）、Scoped（作用域）
- 🔧 **构造函数注册**：自动解析依赖参数
- 📦 **实例注册**：直接注册已创建的对象
- 🎯 **接口和具体类型注册**：支持接口类型和具体类型注册
- 🏷️ **命名注册**：支持同一类型的多个实例注册 ⭐ 新功能
- 🔍 **泛型支持**：类型安全的 `Get[T]()` 和 `MustGet[T]()` 方法
- 🌐 **引用类型支持**：完整支持切片、映射、数组
- 🔒 **线程安全**：所有操作并发安全
- 🛡️ **循环依赖检测**：自动检测并报错
- 📝 **详细错误信息**：清晰的错误提示

## 📦 安装

```bash
go get github.com/yourusername/gofac
```

## 🚀 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "gofac/di"
)

type UserRepo struct {
    ConnStr string
}

func NewUserRepo() *UserRepo {
    return &UserRepo{ConnStr: "localhost:5432"}
}

type UserService struct {
    Repo *UserRepo
}

func NewUserService(repo *UserRepo) *UserService {
    return &UserService{Repo: repo}
}

func main() {
    // 创建容器
    container := di.NewContainer()

    // 注册服务
    container.MustRegister(NewUserRepo, di.Singleton)
    container.MustRegister(NewUserService, di.Transient)

    // 解析服务
    var service *UserService
    container.MustResolve(&service)

    fmt.Println(service.Repo.ConnStr) // 输出: localhost:5432
}
```

### 使用泛型方法

```go
// 使用全局容器
di.MustRegister(NewUserRepo, di.Singleton)
di.MustRegister(NewUserService, di.Transient)

// 泛型解析
service := di.MustGet[*UserService]()
fmt.Println(service.Repo.ConnStr)
```

## 📚 核心概念

### 生命周期

| 生命周期 | 说明 | 使用场景 |
|---------|------|---------|
| **Transient** | 每次解析创建新实例 | 无状态服务、轻量对象 |
| **Singleton** | 全局唯一实例 | 数据库连接、配置对象 |
| **Scoped** | 作用域内唯一 | HTTP 请求上下文、事务 |

### 注册方式

#### 1. 构造函数注册

```go
func NewUserRepo() *UserRepo {
    return &UserRepo{}
}

container.MustRegister(NewUserRepo, di.Singleton)
```

#### 2. 接口和具体类型注册 ⭐ 新功能

**接口注册：**

```go
type ILogger interface {
    Log(msg string)
}

type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(msg string) {
    fmt.Println(msg)
}

func NewConsoleLogger() *ConsoleLogger {
    return &ConsoleLogger{}
}

// 注册为接口类型
container.MustRegisterAs(NewConsoleLogger, (*ILogger)(nil), di.Singleton)

// 通过接口解析
logger := di.MustGet[ILogger]()
logger.Log("Hello")
```

**具体类型注册：**

```go
type UserService struct {
    Name string
}

func NewUserService() *UserService {
    return &UserService{Name: "service"}
}

// 注册为具体类型 *UserService
container.MustRegisterAs(NewUserService, (*UserService)(nil), di.Singleton)

// 通过具体类型解析
service := di.MustGet[*UserService]()
```

> 详细说明请参考 [CONCRETE_TYPE_SUPPORT.md](CONCRETE_TYPE_SUPPORT.md)

#### 3. 实例注册

```go
// 直接注册已创建的实例
config := &Config{AppName: "MyApp", Port: 8080}
container.MustRegisterInstance(config, di.Singleton)

// 解析
resolvedConfig := di.MustGet[*Config]()
```

#### 4. 命名注册 ⭐ 新功能

支持同一类型的多个实例注册，适用于多数据库、多消息队列等场景。

```go
type Database struct {
    Host string
    Port int
}

// 注册多个数据库连接
primary := &Database{Host: "primary.db", Port: 5432}
replica := &Database{Host: "replica.db", Port: 5433}

container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
container.MustRegisterInstanceNamed("replica", replica, di.Singleton)

// 通过名称解析特定实例
var primaryDB *Database
container.MustResolveNamed("primary", &primaryDB)

// 解析所有同类型的实例
var allDBs []*Database
container.MustResolveAll(&allDBs)
fmt.Printf("Total databases: %d\n", len(allDBs)) // 输出: 2
```

> 详细说明请参考 [NAMED_REGISTRATION.md](NAMED_REGISTRATION.md)

### 引用类型支持 ⭐ 新功能

#### 切片（Slice）

```go
// 注册切片
roles := []string{"admin", "user", "guest"}
container.MustRegisterInstance(roles, di.Singleton)

// 作为依赖注入
type UserService struct {
    AllowedRoles []string
}

func NewUserService(roles []string) *UserService {
    return &UserService{AllowedRoles: roles}
}

container.MustRegister(NewUserService, di.Singleton)
```

#### 映射（Map）

```go
// 注册 map
settings := map[string]string{
    "db_host": "localhost",
    "db_port": "5432",
}
container.MustRegisterInstance(settings, di.Singleton)

// 作为依赖注入
type ConfigService struct {
    Settings map[string]string
}

func NewConfigService(settings map[string]string) *ConfigService {
    return &ConfigService{Settings: settings}
}

container.MustRegister(NewConfigService, di.Singleton)
```

#### 数组（Array）

```go
// 注册数组
priorities := [5]int{1, 2, 3, 4, 5}
container.MustRegisterInstance(priorities, di.Singleton)

// 解析
resolved := di.MustGet[[5]int]()
```

### 作用域（Scope）

```go
// 注册 Scoped 服务
container.MustRegister(NewRequestContext, di.Scoped)

// 创建作用域
scope1 := container.NewScope()
scope2 := container.NewScope()

// 每个作用域有独立的实例
ctx1 := di.ScopeMustGet[*RequestContext](scope1)
ctx2 := di.ScopeMustGet[*RequestContext](scope2)

fmt.Println(ctx1 != ctx2) // true
```

## 📖 API 参考

### 注册方法

| 方法 | 说明 | 返回错误 |
|------|------|---------|
| `Register(ctor, scope)` | 构造函数注册 | ✅ |
| `RegisterAs(ctor, iface, scope)` | 构造函数接口注册 | ✅ |
| `RegisterInstance(instance, scope)` | 实例注册 | ✅ |
| `RegisterInstanceAs(instance, iface, scope)` | 实例接口注册 | ✅ |
| `MustRegister(ctor, scope)` | 构造函数注册（panic） | ❌ |
| `MustRegisterAs(ctor, iface, scope)` | 构造函数接口注册（panic） | ❌ |
| `MustRegisterInstance(instance, scope)` | 实例注册（panic） | ❌ |
| `MustRegisterInstanceAs(instance, iface, scope)` | 实例接口注册（panic） | ❌ |

### 解析方法

| 方法 | 说明 | 返回错误 |
|------|------|---------|
| `Resolve(out)` | 指针解析 | ✅ |
| `MustResolve(out)` | 指针解析（panic） | ❌ |
| `Get[T]()` | 泛型解析 | ✅ |
| `MustGet[T]()` | 泛型解析（panic） | ❌ |
| `ScopeGet[T](scope)` | 作用域泛型解析 | ✅ |
| `ScopeMustGet[T](scope)` | 作用域泛型解析（panic） | ❌ |

### 全局容器方法

```go
di.MustRegister(ctor, scope)
di.MustRegisterAs(ctor, iface, scope)
di.MustRegisterInstance(instance, scope)
di.MustRegisterInstanceAs(instance, iface, scope)
di.MustResolve(out)
di.Get[T]()
di.MustGet[T]()
di.GlobalNewScope()
di.ScopeGet[T](scope)
di.ScopeMustGet[T](scope)
di.GlobalReset()
```

## 🎯 使用场景

### Web 应用

```go
// 注册数据库连接（Singleton）
di.MustRegister(NewDatabase, di.Singleton)

// 注册仓储（Singleton）
di.MustRegisterAs(NewUserRepo, (*IUserRepo)(nil), di.Singleton)

// 注册服务（Transient）
di.MustRegister(NewUserService, di.Transient)

// HTTP 处理器
func UserHandler(w http.ResponseWriter, r *http.Request) {
    // 创建请求作用域
    scope := di.GlobalNewScope()

    // 注册请求上下文（Scoped）
    ctx := &RequestContext{RequestID: uuid.New().String()}
    scope.MustRegisterInstance(ctx, di.Scoped)

    // 解析服务
    service := di.ScopeMustGet[*UserService](scope)
    // ... 处理请求
}
```

### 测试

```go
func TestUserService(t *testing.T) {
    container := di.NewContainer()

    // 注入 mock 对象
    mockRepo := &MockUserRepo{}
    container.MustRegisterInstanceAs(mockRepo, (*IUserRepo)(nil), di.Singleton)

    // 注册待测试服务
    container.MustRegister(NewUserService, di.Transient)

    // 测试
    service := di.MustGet[*UserService]()
    // ... 断言
}
```

## 📝 完整示例

查看 `example_demo.go` 获取完整示例代码，包括：

1. RegisterInstance 基础用法
2. RegisterInstanceAs 接口注册
3. Scoped 实例注册
4. 切片类型支持
5. Map 类型支持
6. 数组类型支持
7. 全局容器便捷方法
8. 复杂引用类型组合

运行示例：
```bash
go run example_demo.go
```

## 📚 文档

- [完整特性文档](FEATURES.md) - 详细的 API 参考和使用指南
- [实现总结](IMPLEMENTATION_SUMMARY.md) - 技术实现细节

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行 di 包测试
go test ./di -v

# 运行示例
go run example_demo.go
```

## ⚠️ 注意事项

### 1. 实例注册不支持 Transient

```go
// ❌ 错误：实例注册不支持 Transient
config := &Config{}
container.RegisterInstance(config, di.Transient) // 返回 ErrTransientInstance
```

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

### 3. 循环依赖

```go
// ❌ 错误：循环依赖
func NewA(b *B) *A { return &A{B: b} }
func NewB(a *A) *B { return &B{A: a} }

container.MustRegister(NewA, di.Singleton)
container.MustRegister(NewB, di.Singleton)

// 解析时会报错：ErrResolveCircularDependency
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

本项目受 [Autofac](https://autofac.org/) 启发。

---

**作者**: Your Name
**版本**: v1.1.0
**更新日期**: 2026-02-02

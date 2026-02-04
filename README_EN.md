# Gofac - Go Dependency Injection Container

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[中文](README.md) | [English](README_EN.md)

Gofac is a Go dependency injection (DI) container inspired by [Autofac](https://autofac.org/), providing clean and type-safe dependency management.

## ✨ Features

- 🚀 **Three Lifetimes**: Transient, Singleton, Scoped
- 🔧 **Constructor Registration**: Automatic dependency resolution
- 📦 **Instance Registration**: Direct object registration
- 🎯 **Interface & Concrete Type Registration**: Full type support
- 🏷️ **Named Registration**: Multiple instances of same type
- 🔄 **Slice Auto-Injection**: Automatic collection injection ⭐
- 🗺️ **Map Auto-Injection**: Named instance map creation ⭐
- 🔍 **Generic Support**: Type-safe `Get[T]()` and `MustGet[T]()`
- 🌐 **Reference Types**: Slices, maps, arrays support
- 🔒 **Thread-Safe**: Concurrency-safe operations
- 🛡️ **Circular Dependency Detection**: Automatic detection
- 📝 **Clear Error Messages**: Detailed error information

## 📦 Installation

```bash
go get github.com/yourusername/gofac
```

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "gofac"
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
}

func NewUserService(repo *UserRepo) *UserService {
    return &UserService{Repo: repo}
}

func main() {
    // Create container
    container := gofac.NewContainer()

    // Register services
    container.MustRegister(NewUserRepo, gofac.Singleton)
    container.MustRegister(NewUserService, gofac.Transient)

    // Resolve service
    var service *UserService
    container.MustResolve(&service)

    fmt.Println(service.Repo.ConnStr) // Output: localhost:5432
}
```

### Using Generics

```go
// Use global container
gofac.MustRegister(NewUserRepo, gofac.Singleton)
gofac.MustRegister(NewUserService, gofac.Transient)

// Generic resolution
service := gofac.MustGet[*UserService]()
fmt.Println(service.Repo.ConnStr)
```

## 📚 Core Concepts

### Lifetimes

| Lifetimes | Description | Use Cases |
|---------|------|---------|
| **Transient** | New instance each time | Stateless services, lightweight objects |
| **Singleton** | Global unique instance | DB connections, config objects |
| **Scoped** | Unique within scope | HTTP request context, transactions |

### Registration Methods

#### 1. Constructor Registration

```go
func NewUserRepo() *UserRepo {
    return &UserRepo{}
}

container.MustRegister(NewUserRepo, gofac.Singleton)
```

#### 2. Interface and Concrete Type Registration ⭐ New Feature

**Interface Registration:**

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

// Register as interface type
container.MustRegisterAs(NewConsoleLogger, (*ILogger)(nil), gofac.Singleton)

// Resolve through interface
logger := gofac.MustGet[ILogger]()
logger.Log("Hello")
```

**Concrete Type Registration:**

```go
type UserService struct {
    Name string
}

func NewUserService() *UserService {
    return &UserService{Name: "service"}
}

// Register as concrete type *UserService
container.MustRegisterAs(NewUserService, (*UserService)(nil), gofac.Singleton)

// Resolve through concrete type
service := gofac.MustGet[*UserService]()
```

> See for details [CONCRETE_TYPE_SUPPORT.md](docs/CONCRETE_TYPE_SUPPORT.md)

#### 3. Instance Registration

```go
// Directly register created instance
config := &Config{AppName: "MyApp", Port: 8080}
container.MustRegisterInstance(config, gofac.Singleton)

// Resolve
resolvedConfig := gofac.MustGet[*Config]()
```

#### 4. Named Registration ⭐ New Feature

支持同一类型的多个Instance Registration，适用于多数据库、多消息队列等场景。

```go
type Database struct {
    Host string
    Port int
}

// Register multiple database connections
primary := &Database{Host: "primary.db", Port: 5432}
replica := &Database{Host: "replica.db", Port: 5433}

container.MustRegisterInstanceNamed("primary", primary, gofac.Singleton)
container.MustRegisterInstanceNamed("replica", replica, gofac.Singleton)

// Resolve specific instance by name
var primaryDB *Database
container.MustResolveNamed("primary", &primaryDB)

// Resolve all instances of same type
var allDBs []*Database
container.MustResolveAll(&allDBs)
fmt.Printf("Total databases: %d\n", len(allDBs)) // Output: 2
```

> See for details [NAMED_REGISTRATION.md](docs/NAMED_REGISTRATION.md)

#### 5. Slice Auto-Injection ⭐ New Feature

When a constructor requires a slice parameter, the container intelligently handles it:
- If slice type is registered, use it directly
- If not registered, automatically collect all instances of that element type

```go
type DatabaseManager struct {
    Databases []*Database
}

func NewDatabaseManager(dbs []*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}

// Register multiple database instances
container.MustRegisterInstanceNamed("primary", &Database{Host: "primary"}, gofac.Singleton)
container.MustRegisterInstanceNamed("replica", &Database{Host: "replica"}, gofac.Singleton)

// Register DatabaseManager - automatically injects all *Database instances
container.MustRegister(NewDatabaseManager, gofac.Singleton)

var manager *DatabaseManager
container.MustResolve(&manager)
fmt.Printf("Total databases: %d\n", len(manager.Databases)) // Output: 2
```

> See for details [SLICE_AUTO_INJECTION.md](docs/SLICE_AUTO_INJECTION.md)

#### 6. Map Auto-Injection ⭐ New Feature

When a constructor requires a `map[string]T` parameter, the container intelligently handles it:
- If Map type is registered, use it directly
- If not registered，auto collects all instances of that element type

```go
type CacheManager struct {
    Caches map[string]ICache
}

func NewCacheManager(caches map[string]ICache) *CacheManager {
    return &CacheManager{Caches: caches}
}

// Register multiple cache implementations
container.MustRegisterInstanceAsNamed("redis", &RedisCache{}, (*ICache)(nil), gofac.Singleton)
container.MustRegisterInstanceAsNamed("memory", &MemoryCache{}, (*ICache)(nil), gofac.Singleton)

// Register CacheManager - automatically injects all named cache instances
container.MustRegister(NewCacheManager, gofac.Singleton)

var manager *CacheManager
container.MustResolve(&manager)
fmt.Println(manager.Caches["redis"].Get("key")) // Access by name
fmt.Printf("Total caches: %d\n", len(manager.Caches)) // Output: 2
```

> See for details [MAP_AUTO_INJECTION.md](docs/MAP_AUTO_INJECTION.md)

### Reference Type Support

#### Slice

```go
// Register slice
roles := []string{"admin", "user", "guest"}
container.MustRegisterInstance(roles, gofac.Singleton)

// Inject as dependency
type UserService struct {
    AllowedRoles []string
}

func NewUserService(roles []string) *UserService {
    return &UserService{AllowedRoles: roles}
}

container.MustRegister(NewUserService, gofac.Singleton)
```

#### Map

```go
// Register map
settings := map[string]string{
    "db_host": "localhost",
    "db_port": "5432",
}
container.MustRegisterInstance(settings, gofac.Singleton)

// Inject as dependency
type ConfigService struct {
    Settings map[string]string
}

func NewConfigService(settings map[string]string) *ConfigService {
    return &ConfigService{Settings: settings}
}

container.MustRegister(NewConfigService, gofac.Singleton)
```

#### Array

```go
// Register array
priorities := [5]int{1, 2, 3, 4, 5}
container.MustRegisterInstance(priorities, gofac.Singleton)

// Resolve
resolved := gofac.MustGet[[5]int]()
```

### Scopes

```go
// Register Scoped service
container.MustRegister(NewRequestContext, gofac.Scoped)

// Create scopes
scope1 := container.NewScope()
scope2 := container.NewScope()

// Each scope has independent instances
ctx1 := gofac.ScopeMustGet[*RequestContext](scope1)
ctx2 := gofac.ScopeMustGet[*RequestContext](scope2)

fmt.Println(ctx1 != ctx2) // true
```

## 📖 API Reference

### Registration Methods

| Method | Description | Return Error |
|------|------|---------|
| `Register(ctor, scope)` | Constructor Registration | ✅ |
| `RegisterAs(ctor, iface, scope)` | Constructor interface registration | ✅ |
| `RegisterInstance(instance, scope)` | Instance Registration | ✅ |
| `RegisterInstanceAs(instance, iface, scope)` | Instance interface registration | ✅ |
| `MustRegister(ctor, scope)` | Constructor Registration（panic） | ❌ |
| `MustRegisterAs(ctor, iface, scope)` | Constructor interface registration（panic） | ❌ |
| `MustRegisterInstance(instance, scope)` | Instance Registration（panic） | ❌ |
| `MustRegisterInstanceAs(instance, iface, scope)` | Instance interface registration（panic） | ❌ |

### Resolution Methods

| Method                   | Description | Return Error |
|--------------------------|------|--------------|
| `Resolve(out)`           | Pointer resolution | ✅            |
| `MustResolve(out)`       | Pointer resolution（panic） | ❌            |
| `Get[T]()`               | Generic resolution | ✅            |
| `MustGet[T]()`           | Generic resolution（panic） | ❌            |
| `ScopeGet[T](scope)`     | 作用域Generic resolution | ✅            |
| `ScopeMustGet[T](scope)` | 作用域Generic resolution（panic） | ❌            |

### Global Container Methods

```go
gofac.MustRegister(ctor, scope)
gofac.MustRegisterAs(ctor, iface, scope)
gofac.MustRegisterInstance(instance, scope)
gofac.MustRegisterInstanceAs(instance, iface, scope)
gofac.MustResolve(out)
gofac.Get[T]()
gofac.MustGet[T]()
gofac.GlobalNewScope()
gofac.ScopeGet[T](scope)
gofac.ScopeMustGet[T](scope)
gofac.GlobalReset()
```

## 🎯 Use Cases

### Web Application

```go
// Register database connection (Singleton)
gofac.MustRegister(NewDatabase, gofac.Singleton)

// Register repository (Singleton)
gofac.MustRegisterAs(NewUserRepo, (*IUserRepo)(nil), gofac.Singleton)

// Register services（Transient）
gofac.MustRegister(NewUserService, gofac.Transient)

// HTTP handler
func UserHandler(w http.ResponseWriter, r *http.Request) {
    // Create request scope
    scope := gofac.GlobalNewScope()

    // Register request context (Scoped)
    ctx := &RequestContext{RequestID: uuid.New().String()}
    scope.MustRegisterInstance(ctx, gofac.Scoped)

    // Resolve service
    service := gofac.ScopeMustGet[*UserService](scope)
    // ... Handle Request
}
```

### Testing

```go
func TestUserService(t *testing.T) {
    container := gofac.NewContainer()

    // Inject mock object
    mockRepo := &MockUserRepo{}
    container.MustRegisterInstanceAs(mockRepo, (*IUserRepo)(nil), gofac.Singleton)

    // Register Testing Service
    container.MustRegister(NewUserService, gofac.Transient)

    // Testing
    service := gofac.MustGet[*UserService]()
    // ... Assert
}
```

## 📝 Complete Examples

View `example_demo.go` TO Get Complete Examples' Code，Includes：

1. RegisterInstance basic usage
2. RegisterInstanceAs Interface Registration
3. Scoped Instance Registration
4. Slice type support
5. Map type support
6. Array type support
7. Global container convenience methods
8. Complex reference type combinations

Run example:
```bash
go run example_demo.go
```

## 📚 Documentation

- [Entry Documentation](docs/FEATURES.md) - More Detail API Reference And Usage
- [Implementation Summary](docs/IMPLEMENTATION_SUMMARY.md) - Technical implementation details

## 🧪 Testing

```bash
# Run All Testing
go test ./...

# Run example
go run example_demo.go
```

## ⚠️ Important Notes

### 1. Instance Registration Unsupported Transient

```go
// ❌ 错误：Instance Registration不支持 Transient
config := &Config{}
container.RegisterInstance(config, gofac.Transient) // Returns ErrTransientInstance
```

### 2. Concurrency Safety of Reference Types

```go
// ❌ Unsafe: Multiple goroutines modifying simultaneously
settings := map[string]string{"key": "value"}
container.MustRegisterInstance(settings, gofac.Singleton)

// ✅ Safe: Use sync.Map
var settings sync.Map
container.MustRegisterInstance(&settings, gofac.Singleton)

// ✅ Safe: Read-only access
roles := []string{"admin", "user"}  // Don't modify after registration
container.MustRegisterInstance(roles, gofac.Singleton)
```

### 3. Circular Dependencies

```go
// ❌ 错误：Circular Dependencies
func NewA(b *B) *A { return &A{B: b} }
func NewB(a *A) *B { return &B{A: a} }

container.MustRegister(NewA, gofac.Singleton)
container.MustRegister(NewB, gofac.Singleton)

// Will error during resolution: ErrResolveCircularDependency
```

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License

## 🙏 Acknowledgments

This project is inspired by [Autofac](https://autofac.org/).

---

**Author**: Ngone6325
**Version**: v1.1.0
**Updated**: 2026-02-02

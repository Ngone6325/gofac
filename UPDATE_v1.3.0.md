# Gofac v1.3.0 更新说明 - 命名注册功能

## 更新概述

本次更新添加了**命名注册功能**，允许为同一类型注册多个实例，并通过名称来区分和解析它们。同时补充了大量实际项目场景的示例。

## 新增功能

### 1. 命名注册 API

#### 注册方法

```go
// 命名实例注册
func (c *Container) RegisterInstanceNamed(name string, instance any, scope LifetimeScope) error
func (c *Container) MustRegisterInstanceNamed(name string, instance any, scope LifetimeScope)

// 命名实例接口注册
func (c *Container) RegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope) error
func (c *Container) MustRegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope)
```

#### 解析方法

```go
// 通过名称解析特定实例
func (c *Container) ResolveNamed(name string, out any) error
func (c *Container) MustResolveNamed(name string, out any)

// 解析所有同类型的实例
func (c *Container) ResolveAll(out any) error
func (c *Container) MustResolveAll(out any)
```

### 2. 核心特性

- ✅ **支持同一类型的多个实例**：通过名称区分
- ✅ **命名解析**：通过名称解析特定实例
- ✅ **批量解析**：解析所有同类型的实例（包括默认和命名实例）
- ✅ **接口支持**：可以将实例注册为接口类型
- ✅ **生命周期支持**：Singleton 和 Scoped
- ✅ **线程安全**：所有操作并发安全

## 使用场景

### 场景 1：多数据库连接

```go
// 注册主库和从库
primary := &Database{Host: "primary.db", Port: 5432}
replica1 := &Database{Host: "replica1.db", Port: 5432}
replica2 := &Database{Host: "replica2.db", Port: 5432}

container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
container.MustRegisterInstanceNamed("replica1", replica1, di.Singleton)
container.MustRegisterInstanceNamed("replica2", replica2, di.Singleton)

// 读写分离
var writeDB *Database
container.MustResolveNamed("primary", &writeDB)

var readDBs []*Database
container.MustResolveAll(&readDBs)
```

### 场景 2：消息队列

```go
// 注册多个消息队列
orderQueue := &MessageQueue{Topic: "orders"}
paymentQueue := &MessageQueue{Topic: "payments"}
notificationQueue := &MessageQueue{Topic: "notifications"}

container.MustRegisterInstanceNamed("order", orderQueue, di.Singleton)
container.MustRegisterInstanceNamed("payment", paymentQueue, di.Singleton)
container.MustRegisterInstanceNamed("notification", notificationQueue, di.Singleton)

// 使用特定队列
var orderMQ *MessageQueue
container.MustResolveNamed("order", &orderMQ)
orderMQ.Publish("New order")

// 广播到所有队列
var allQueues []*MessageQueue
container.MustResolveAll(&allQueues)
for _, queue := range allQueues {
    queue.Publish("System message")
}
```

### 场景 3：多种缓存实现

```go
// 注册多种缓存
redis := &RedisCache{Name: "MainRedis"}
memory := &MemoryCache{Name: "LocalMemory"}
sessionRedis := &RedisCache{Name: "SessionRedis"}

container.MustRegisterInstanceAsNamed("redis", redis, (*ICache)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("memory", memory, (*ICache)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("session", sessionRedis, (*ICache)(nil), di.Singleton)

// 使用特定缓存
var mainCache ICache
container.MustResolveNamed("redis", &mainCache)

// 获取所有缓存
var allCaches []ICache
container.MustResolveAll(&allCaches)
```

### 场景 4：微服务客户端

```go
// 注册多个微服务客户端
userService := &ServiceClient{BaseURL: "http://user-service:8080"}
orderService := &ServiceClient{BaseURL: "http://order-service:8081"}
paymentService := &ServiceClient{BaseURL: "http://payment-service:8082"}

container.MustRegisterInstanceNamed("user", userService, di.Singleton)
container.MustRegisterInstanceNamed("order", orderService, di.Singleton)
container.MustRegisterInstanceNamed("payment", paymentService, di.Singleton)

// 调用特定服务
var userClient *ServiceClient
container.MustResolveNamed("user", &userClient)
userClient.Call("/api/users/1001")

// 健康检查所有服务
var allClients []*ServiceClient
container.MustResolveAll(&allClients)
for _, client := range allClients {
    client.Call("/health")
}
```

## 技术实现

### 数据结构

```go
type Container struct {
    services      map[reflect.Type]*ServiceDef            // 默认（无名）服务
    namedServices map[string]map[reflect.Type]*ServiceDef // 命名服务
    mu            sync.RWMutex
}
```

### 关键特性

1. **独立存储**：命名服务和默认服务分开存储
2. **名称唯一性**：同一名称和类型的组合只能注册一次
3. **批量解析**：`ResolveAll` 同时返回默认和所有命名实例
4. **线程安全**：使用读写锁保护并发访问

## 新增示例

在 `main.go` 中添加了 5 个实际项目场景的示例：

- **示例 9**：命名注册 - 多个同类型实例
- **示例 10**：多数据库连接（实际项目场景）
- **示例 11**：消息队列（实际项目场景）
- **示例 12**：缓存策略（实际项目场景）
- **示例 13**：微服务客户端（实际项目场景）

运行示例：
```bash
go run main.go
```

## 测试覆盖

新增测试文件：`di/named_test.go`

包含 8 个测试用例：

1. ✅ `TestRegisterInstanceNamed` - 命名实例注册
2. ✅ `TestResolveAll` - 解析所有实例
3. ✅ `TestRegisterInstanceNamed_DuplicateName` - 重复名称检测
4. ✅ `TestRegisterInstanceNamed_EmptyName` - 空名称检测
5. ✅ `TestRegisterInstanceAsNamed` - 接口命名注册
6. ✅ `TestResolveAll_Interface` - 解析所有接口实现
7. ✅ `TestRealWorldScenario_MultipleQueues` - 实际场景测试

所有测试通过：
```
PASS
ok      gofac/di        0.008s
ok      gofac/tests     0.008s
```

## 文档更新

### 新增文档

- **NAMED_REGISTRATION.md** - 命名注册功能的详细文档
  - 核心 API
  - 使用示例
  - 特性说明
  - 使用场景
  - 最佳实践

### 更新文档

- **README.md** - 更新特性列表，添加命名注册示例

## 向后兼容性

✅ **完全向后兼容**

- 所有原有代码无需修改
- 命名注册作为补充功能
- 不影响现有的注册和解析逻辑

## API 变化

### 新增 API（8 个方法）

**注册方法：**
- `RegisterInstanceNamed`
- `MustRegisterInstanceNamed`
- `RegisterInstanceAsNamed`
- `MustRegisterInstanceAsNamed`

**解析方法：**
- `ResolveNamed`
- `MustResolveNamed`
- `ResolveAll`
- `MustResolveAll`

### 无破坏性变更

所有现有 API 保持不变，新功能作为补充。

## 使用建议

### 何时使用命名注册

**推荐场景：**
- ✅ 多个数据库连接（主库、从库）
- ✅ 多个消息队列（不同主题）
- ✅ 多种缓存实现（Redis、Memory）
- ✅ 多个微服务客户端
- ✅ 插件系统
- ✅ 中间件链
- ✅ 配置管理（多环境）

**不推荐场景：**
- ❌ 只有一个实例的类型（使用默认注册）
- ❌ 需要依赖注入的服务（命名服务不参与自动注入）

### 最佳实践

1. **使用常量定义名称**
```go
const (
    DBPrimary  = "primary"
    DBReplica1 = "replica1"
)
```

2. **封装解析逻辑**
```go
type DBManager struct {
    container *di.Container
}

func (m *DBManager) GetPrimary() *Database {
    var db *Database
    m.container.MustResolveNamed("primary", &db)
    return db
}
```

3. **使用工厂模式**
```go
type CacheFactory struct {
    container *di.Container
}

func (f *CacheFactory) GetCache(name string) ICache {
    var cache ICache
    f.container.MustResolveNamed(name, &cache)
    return cache
}
```

## 限制和注意事项

1. **命名注册目前只支持实例注册**，不支持构造函数注册
2. **命名服务不参与依赖注入**：不能作为构造函数参数自动注入
3. **ResolveAll 返回顺序不保证**：不要依赖返回顺序
4. **名称必须唯一**：同一名称和类型的组合不能重复注册

## 性能影响

- ✅ **最小性能开销**：命名服务使用独立的 map 存储
- ✅ **线程安全**：使用读写锁，读操作不互斥
- ✅ **内存效率**：只在需要时创建命名服务 map

## 版本信息

- **版本号**: v1.3.0
- **发布日期**: 2026-02-02
- **兼容性**: 完全向后兼容 v1.2.0

## 总结

本次更新通过添加命名注册功能，使 gofac 能够更好地支持实际项目中的复杂场景。主要改进：

- ✅ 支持同一类型的多个实例注册
- ✅ 提供命名解析和批量解析功能
- ✅ 补充了 5 个实际项目场景的示例
- ✅ 完整的测试覆盖（8 个测试用例）
- ✅ 详细的文档说明
- ✅ 完全向后兼容

---

**作者**: Gofac Team
**更新时间**: 2026-02-02

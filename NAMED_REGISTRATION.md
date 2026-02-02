# 命名注册功能文档

## 概述

命名注册功能允许你为同一类型注册多个实例，通过名称来区分和解析它们。这在实际项目中非常有用，例如：

- 多个数据库连接（主库、从库）
- 多个消息队列（不同的主题）
- 多种缓存实现（Redis、Memory）
- 多个微服务客户端
- 插件系统

## 核心 API

### 注册方法

```go
// 命名实例注册
func (c *Container) RegisterInstanceNamed(name string, instance any, scope LifetimeScope) error
func (c *Container) MustRegisterInstanceNamed(name string, instance any, scope LifetimeScope)

// 命名实例接口注册
func (c *Container) RegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope) error
func (c *Container) MustRegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope)
```

### 解析方法

```go
// 通过名称解析特定实例
func (c *Container) ResolveNamed(name string, out any) error
func (c *Container) MustResolveNamed(name string, out any)

// 解析所有同类型的实例（包括默认和所有命名实例）
func (c *Container) ResolveAll(out any) error
func (c *Container) MustResolveAll(out any)
```

## 使用示例

### 示例 1：多数据库连接

```go
type DBConnection struct {
    Name string
    Host string
    Port int
}

func main() {
    container := di.NewContainer()

    // 注册主库和多个从库
    primary := &DBConnection{Name: "primary", Host: "primary.db.com", Port: 5432}
    replica1 := &DBConnection{Name: "replica1", Host: "replica1.db.com", Port: 5432}
    replica2 := &DBConnection{Name: "replica2", Host: "replica2.db.com", Port: 5432}

    container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
    container.MustRegisterInstanceNamed("replica1", replica1, di.Singleton)
    container.MustRegisterInstanceNamed("replica2", replica2, di.Singleton)

    // 解析主库
    var primaryDB *DBConnection
    container.MustResolveNamed("primary", &primaryDB)
    fmt.Printf("Primary DB: %s:%d\n", primaryDB.Host, primaryDB.Port)

    // 解析所有数据库连接
    var allDBs []*DBConnection
    container.MustResolveAll(&allDBs)
    fmt.Printf("Total DB connections: %d\n", len(allDBs))
}
```

**输出：**
```
Primary DB: primary.db.com:5432
Total DB connections: 3
```

### 示例 2：消息队列

```go
type MessageQueue struct {
    Name  string
    Topic string
}

func (mq *MessageQueue) Publish(message string) {
    fmt.Printf("[%s] Publishing to %s: %s\n", mq.Name, mq.Topic, message)
}

func main() {
    container := di.NewContainer()

    // 注册多个消息队列
    orderQueue := &MessageQueue{Name: "OrderQueue", Topic: "orders"}
    paymentQueue := &MessageQueue{Name: "PaymentQueue", Topic: "payments"}
    notificationQueue := &MessageQueue{Name: "NotificationQueue", Topic: "notifications"}

    container.MustRegisterInstanceNamed("order", orderQueue, di.Singleton)
    container.MustRegisterInstanceNamed("payment", paymentQueue, di.Singleton)
    container.MustRegisterInstanceNamed("notification", notificationQueue, di.Singleton)

    // 使用特定队列
    var orderMQ *MessageQueue
    container.MustResolveNamed("order", &orderMQ)
    orderMQ.Publish("New order #12345")

    // 广播到所有队列
    var allQueues []*MessageQueue
    container.MustResolveAll(&allQueues)
    for _, queue := range allQueues {
        queue.Publish("System maintenance at 2AM")
    }
}
```

**输出：**
```
[OrderQueue] Publishing to orders: New order #12345
[OrderQueue] Publishing to orders: System maintenance at 2AM
[PaymentQueue] Publishing to payments: System maintenance at 2AM
[NotificationQueue] Publishing to notifications: System maintenance at 2AM
```

### 示例 3：多种缓存实现

```go
type ICache interface {
    Get(key string) string
    Set(key string, value string)
}

type RedisCache struct {
    Name string
}

func (r *RedisCache) Get(key string) string {
    return fmt.Sprintf("[Redis:%s] %s", r.Name, key)
}

func (r *RedisCache) Set(key string, value string) {
    // implementation
}

type MemoryCache struct {
    Name string
}

func (m *MemoryCache) Get(key string) string {
    return fmt.Sprintf("[Memory:%s] %s", m.Name, key)
}

func (m *MemoryCache) Set(key string, value string) {
    // implementation
}

func main() {
    container := di.NewContainer()

    // 注册多种缓存实现
    redis := &RedisCache{Name: "MainRedis"}
    memory := &MemoryCache{Name: "LocalMemory"}
    sessionRedis := &RedisCache{Name: "SessionRedis"}

    container.MustRegisterInstanceAsNamed("redis", redis, (*ICache)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("memory", memory, (*ICache)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("session", sessionRedis, (*ICache)(nil), di.Singleton)

    // 使用特定缓存
    var mainCache ICache
    container.MustResolveNamed("redis", &mainCache)
    fmt.Println(mainCache.Get("user:1001"))

    // 获取所有缓存实现
    var allCaches []ICache
    container.MustResolveAll(&allCaches)
    fmt.Printf("All cache implementations: %d\n", len(allCaches))
}
```

**输出：**
```
[Redis:MainRedis] user:1001
All cache implementations: 3
```

### 示例 4：微服务客户端

```go
type ServiceClient struct {
    ServiceName string
    BaseURL     string
}

func (sc *ServiceClient) Call(endpoint string) string {
    return fmt.Sprintf("[%s] Calling %s%s", sc.ServiceName, sc.BaseURL, endpoint)
}

func main() {
    container := di.NewContainer()

    // 注册多个微服务客户端
    userService := &ServiceClient{ServiceName: "UserService", BaseURL: "http://user-service:8080"}
    orderService := &ServiceClient{ServiceName: "OrderService", BaseURL: "http://order-service:8081"}
    paymentService := &ServiceClient{ServiceName: "PaymentService", BaseURL: "http://payment-service:8082"}

    container.MustRegisterInstanceNamed("user", userService, di.Singleton)
    container.MustRegisterInstanceNamed("order", orderService, di.Singleton)
    container.MustRegisterInstanceNamed("payment", paymentService, di.Singleton)

    // 调用特定服务
    var userClient *ServiceClient
    container.MustResolveNamed("user", &userClient)
    fmt.Println(userClient.Call("/api/users/1001"))

    // 健康检查所有服务
    var allClients []*ServiceClient
    container.MustResolveAll(&allClients)
    for _, client := range allClients {
        fmt.Println(client.Call("/health"))
    }
}
```

**输出：**
```
[UserService] Calling http://user-service:8080/api/users/1001
[UserService] Calling http://user-service:8080/health
[OrderService] Calling http://order-service:8081/health
[PaymentService] Calling http://payment-service:8082/health
```

## 特性说明

### 1. 命名注册

- **支持同一类型的多个实例**：通过名称区分
- **名称不能为空**：空名称会返回错误
- **名称不能重复**：同一名称和类型的组合只能注册一次
- **支持接口类型**：可以将实例注册为接口类型

### 2. 命名解析

- **ResolveNamed**：通过名称解析特定的实例
- **ResolveAll**：解析所有同类型的实例（包括默认和所有命名实例）

### 3. 生命周期支持

- ✅ **Singleton**：全局唯一实例
- ✅ **Scoped**：作用域内唯一实例
- ❌ **Transient**：不支持（实例已创建，无法每次返回新实例）

### 4. 与默认注册的关系

- 命名注册和默认注册是独立的
- `ResolveAll` 会同时返回默认实例和所有命名实例
- 可以同时使用默认注册和命名注册

## 使用场景

### 1. 多数据源

```go
// 主库用于写操作
container.MustRegisterInstanceNamed("primary", primaryDB, di.Singleton)

// 从库用于读操作
container.MustRegisterInstanceNamed("replica1", replica1DB, di.Singleton)
container.MustRegisterInstanceNamed("replica2", replica2DB, di.Singleton)

// 读写分离
var writeDB *Database
container.MustResolveNamed("primary", &writeDB)

var readDBs []*Database
container.MustResolveAll(&readDBs)
// 从 readDBs 中随机选择一个用于读操作
```

### 2. 配置管理

```go
// 不同环境的配置
container.MustRegisterInstanceNamed("dev", devConfig, di.Singleton)
container.MustRegisterInstanceNamed("test", testConfig, di.Singleton)
container.MustRegisterInstanceNamed("prod", prodConfig, di.Singleton)

// 根据环境变量选择配置
env := os.Getenv("ENV")
var config *Config
container.MustResolveNamed(env, &config)
```

### 3. 插件系统

```go
// 注册多个插件
container.MustRegisterInstanceAsNamed("plugin1", plugin1, (*IPlugin)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("plugin2", plugin2, (*IPlugin)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("plugin3", plugin3, (*IPlugin)(nil), di.Singleton)

// 加载所有插件
var plugins []IPlugin
container.MustResolveAll(&plugins)
for _, plugin := range plugins {
    plugin.Initialize()
}
```

### 4. 中间件链

```go
// 注册多个中间件
container.MustRegisterInstanceAsNamed("auth", authMiddleware, (*IMiddleware)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("logging", loggingMiddleware, (*IMiddleware)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("cors", corsMiddleware, (*IMiddleware)(nil), di.Singleton)

// 构建中间件链
var middlewares []IMiddleware
container.MustResolveAll(&middlewares)
chain := buildMiddlewareChain(middlewares)
```

## 注意事项

1. **命名注册目前只支持实例注册**，不支持构造函数注册
2. **名称必须唯一**：同一名称和类型的组合不能重复注册
3. **ResolveAll 返回的顺序不保证**：不要依赖返回顺序
4. **命名服务不参与依赖注入**：命名服务不能作为构造函数的参数自动注入

## 最佳实践

### 1. 使用常量定义名称

```go
const (
    DBPrimary  = "primary"
    DBReplica1 = "replica1"
    DBReplica2 = "replica2"
)

container.MustRegisterInstanceNamed(DBPrimary, primaryDB, di.Singleton)
container.MustResolveNamed(DBPrimary, &db)
```

### 2. 封装解析逻辑

```go
type DBManager struct {
    container *di.Container
}

func (m *DBManager) GetPrimary() *Database {
    var db *Database
    m.container.MustResolveNamed("primary", &db)
    return db
}

func (m *DBManager) GetReplicas() []*Database {
    var dbs []*Database
    m.container.MustResolveAll(&dbs)
    // 过滤掉主库
    replicas := make([]*Database, 0)
    for _, db := range dbs {
        if db.Name != "primary" {
            replicas = append(replicas, db)
        }
    }
    return replicas
}
```

### 3. 使用工厂模式

```go
type CacheFactory struct {
    container *di.Container
}

func (f *CacheFactory) GetCache(name string) ICache {
    var cache ICache
    f.container.MustResolveNamed(name, &cache)
    return cache
}

func (f *CacheFactory) GetAllCaches() []ICache {
    var caches []ICache
    f.container.MustResolveAll(&caches)
    return caches
}
```

## 完整示例

完整的示例代码请参考 `main.go` 中的示例 9-13：

- 示例 9：命名注册 - 多个同类型实例
- 示例 10：多数据库连接（实际项目场景）
- 示例 11：消息队列（实际项目场景）
- 示例 12：缓存策略（实际项目场景）
- 示例 13：微服务客户端（实际项目场景）

## 测试

完整的测试用例请参考 `di/named_test.go`，包括：

- ✅ 命名实例注册
- ✅ 命名解析
- ✅ 解析所有实例
- ✅ 重复名称检测
- ✅ 空名称检测
- ✅ 接口命名注册
- ✅ 实际场景测试

---

**版本**: v1.3.0
**更新日期**: 2026-02-02

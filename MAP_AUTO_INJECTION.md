# Map 自动注入功能文档

## 概述

gofac 容器现在支持**Map 类型的自动注入**。当构造函数需要 `map[string]T` 类型的参数时，容器会智能地处理依赖注入：

1. **优先使用已注册的 Map 实例**：如果 Map 类型本身已注册，直接使用
2. **自动收集命名注册实例**：如果 Map 类型未注册，自动收集所有该值类型的命名注册实例，以名称为键创建 Map

## 使用场景

### 场景 1：直接注册 Map（传统方式）

```go
type ConfigService struct {
    Settings map[string]string
}

func NewConfigService(settings map[string]string) *ConfigService {
    return &ConfigService{Settings: settings}
}

func main() {
    container := di.NewContainer()

    // 直接注册 map 实例
    settings := map[string]string{
        "db_host": "localhost",
        "db_port": "5432",
    }
    container.MustRegisterInstance(settings, di.Singleton)

    // 注册依赖 map 的服务
    container.MustRegister(NewConfigService, di.Singleton)

    // 解析服务 - 会注入已注册的 map
    var service *ConfigService
    container.MustResolve(&service)

    fmt.Println(service.Settings["db_host"]) // localhost
}
```

### 场景 2：自动收集命名实例（新功能）

```go
type Database struct {
    Host string
    Port int
}

type DatabaseManager struct {
    Databases map[string]*Database
}

func NewDatabaseManager(dbs map[string]*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}

func main() {
    container := di.NewContainer()

    // 注册多个数据库实例（使用命名注册）
    primary := &Database{Host: "primary.db", Port: 5432}
    replica1 := &Database{Host: "replica1.db", Port: 5433}
    replica2 := &Database{Host: "replica2.db", Port: 5434}

    container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
    container.MustRegisterInstanceNamed("replica1", replica1, di.Singleton)
    container.MustRegisterInstanceNamed("replica2", replica2, di.Singleton)

    // 注册 DatabaseManager
    // 容器会自动收集所有命名的 *Database 实例并作为 map 注入
    container.MustRegister(NewDatabaseManager, di.Singleton)

    // 解析服务
    var manager *DatabaseManager
    container.MustResolve(&manager)

    fmt.Printf("Primary DB: %s:%d\n",
        manager.Databases["primary"].Host,
        manager.Databases["primary"].Port)
    fmt.Printf("Total databases: %d\n", len(manager.Databases)) // 3
}
```

## 工作原理

### 解析逻辑

当容器解析构造函数的依赖参数时：

```
1. 检查参数类型是否为 map[string]T
   ├─ 是 map[string]T 类型
   │  ├─ 检查 map 类型本身是否已注册
   │  │  ├─ 已注册 → 直接解析并注入该 map 实例
   │  │  └─ 未注册 → 自动收集值类型的所有命名注册实例
   │  │     ├─ 遍历所有命名注册
   │  │     ├─ 查找值类型 T 的命名实例
   │  │     ├─ 以名称为键，实例为值创建 map
   │  │     └─ 注入 map
   │  └─ 返回 map
   └─ 非 map[string]T 类型 → 正常解析
```

### 示例流程

```go
// 场景：未注册 map[string]*Database 类型

container.MustRegisterInstanceNamed("primary", &Database{Host: "primary"}, di.Singleton)
container.MustRegisterInstanceNamed("replica", &Database{Host: "replica"}, di.Singleton)

// 当解析 NewDatabaseManager(map[string]*Database) 时：
// 1. 检测到参数类型是 map[string]*Database
// 2. 检查 map[string]*Database 是否已注册 → 未注册
// 3. 提取值类型 *Database
// 4. 收集所有命名的 *Database 实例：
//    - 遍历 namedServices
//    - 找到 "primary" -> *Database
//    - 找到 "replica" -> *Database
// 5. 创建 map[string]*Database{"primary": db1, "replica": db2}
// 6. 注入到构造函数
```

## 完整示例

### 示例 1：多数据库连接管理

```go
type Database struct {
    Host string
    Port int
}

func (d *Database) Connect() string {
    return fmt.Sprintf("Connected to %s:%d", d.Host, d.Port)
}

type DatabaseManager struct {
    Databases map[string]*Database
}

func NewDatabaseManager(dbs map[string]*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}

func (dm *DatabaseManager) GetPrimary() *Database {
    return dm.Databases["primary"]
}

func (dm *DatabaseManager) GetReplica(name string) *Database {
    return dm.Databases[name]
}

func main() {
    container := di.NewContainer()

    // 注册多个数据库连接
    primary := &Database{Host: "primary.db.com", Port: 5432}
    replica1 := &Database{Host: "replica1.db.com", Port: 5432}
    replica2 := &Database{Host: "replica2.db.com", Port: 5432}

    container.MustRegisterInstanceNamed("primary", primary, di.Singleton)
    container.MustRegisterInstanceNamed("replica1", replica1, di.Singleton)
    container.MustRegisterInstanceNamed("replica2", replica2, di.Singleton)

    // 注册管理器 - 自动注入所有数据库
    container.MustRegister(NewDatabaseManager, di.Singleton)

    // 解析并使用
    var manager *DatabaseManager
    container.MustResolve(&manager)

    fmt.Println(manager.GetPrimary().Connect())
    fmt.Println(manager.GetReplica("replica1").Connect())
    fmt.Printf("Total databases: %d\n", len(manager.Databases))
}
```

**输出：**
```
Connected to primary.db.com:5432
Connected to replica1.db.com:5432
Total databases: 3
```

### 示例 2：缓存策略管理

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
    fmt.Printf("[Redis:%s] Set %s=%s\n", r.Name, key, value)
}

type MemoryCache struct {
    Name string
}

func (m *MemoryCache) Get(key string) string {
    return fmt.Sprintf("[Memory:%s] %s", m.Name, key)
}

func (m *MemoryCache) Set(key string, value string) {
    fmt.Printf("[Memory:%s] Set %s=%s\n", m.Name, key, value)
}

type CacheManager struct {
    Caches map[string]ICache
}

func NewCacheManager(caches map[string]ICache) *CacheManager {
    return &CacheManager{Caches: caches}
}

func (cm *CacheManager) GetCache(name string) ICache {
    return cm.Caches[name]
}

func main() {
    container := di.NewContainer()

    // 注册多个缓存实现
    redis := &RedisCache{Name: "MainRedis"}
    memory := &MemoryCache{Name: "LocalMemory"}
    sessionRedis := &RedisCache{Name: "SessionRedis"}

    container.MustRegisterInstanceAsNamed("redis", redis, (*ICache)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("memory", memory, (*ICache)(nil), di.Singleton)
    container.MustRegisterInstanceAsNamed("session", sessionRedis, (*ICache)(nil), di.Singleton)

    // 注册缓存管理器 - 自动注入所有缓存
    container.MustRegister(NewCacheManager, di.Singleton)

    // 解析并使用
    var manager *CacheManager
    container.MustResolve(&manager)

    // 使用不同的缓存
    fmt.Println(manager.GetCache("redis").Get("user:1001"))
    fmt.Println(manager.GetCache("memory").Get("temp:data"))
    fmt.Printf("Total caches: %d\n", len(manager.Caches))
}
```

**输出：**
```
[Redis:MainRedis] user:1001
[Memory:LocalMemory] temp:data
Total caches: 3
```

### 示例 3：微服务客户端管理

```go
type ServiceClient struct {
    Name    string
    BaseURL string
}

func (sc *ServiceClient) Call(endpoint string) string {
    return fmt.Sprintf("[%s] Calling %s%s", sc.Name, sc.BaseURL, endpoint)
}

type ServiceRegistry struct {
    Clients map[string]*ServiceClient
}

func NewServiceRegistry(clients map[string]*ServiceClient) *ServiceRegistry {
    return &ServiceRegistry{Clients: clients}
}

func (sr *ServiceRegistry) GetClient(name string) *ServiceClient {
    return sr.Clients[name]
}

func (sr *ServiceRegistry) HealthCheck() {
    for name, client := range sr.Clients {
        fmt.Printf("%s: %s\n", name, client.Call("/health"))
    }
}

func main() {
    container := di.NewContainer()

    // 注册多个微服务客户端
    userService := &ServiceClient{
        Name:    "UserService",
        BaseURL: "http://user-service:8080",
    }
    orderService := &ServiceClient{
        Name:    "OrderService",
        BaseURL: "http://order-service:8081",
    }
    paymentService := &ServiceClient{
        Name:    "PaymentService",
        BaseURL: "http://payment-service:8082",
    }

    container.MustRegisterInstanceNamed("user", userService, di.Singleton)
    container.MustRegisterInstanceNamed("order", orderService, di.Singleton)
    container.MustRegisterInstanceNamed("payment", paymentService, di.Singleton)

    // 注册服务注册中心 - 自动注入所有客户端
    container.MustRegister(NewServiceRegistry, di.Singleton)

    // 解析并使用
    var registry *ServiceRegistry
    container.MustResolve(&registry)

    // 调用特定服务
    fmt.Println(registry.GetClient("user").Call("/api/users/1001"))

    // 健康检查所有服务
    fmt.Println("\nHealth check:")
    registry.HealthCheck()
}
```

**输出：**
```
[UserService] Calling http://user-service:8080/api/users/1001

Health check:
user: [UserService] Calling http://user-service:8080/health
order: [OrderService] Calling http://order-service:8081/health
payment: [PaymentService] Calling http://payment-service:8082/health
```

## 优先级规则

1. **Map 类型已注册** → 使用已注册的 map 实例
2. **Map 类型未注册** → 自动收集值类型的命名注册实例

这个设计确保了：
- ✅ **向后兼容**：现有代码继续工作
- ✅ **灵活性**：可以选择直接注册 map 或让容器自动收集
- ✅ **智能化**：容器自动处理复杂的依赖关系
- ✅ **类型安全**：只收集匹配值类型的命名实例

## 注意事项

### 1. 空 Map

如果没有注册任何命名实例，会注入空 map：

```go
func NewService(items map[string]*Item) *Service {
    return &Service{Items: items}
}

container.MustRegister(NewService, di.Singleton)

var service *Service
container.MustResolve(&service)

fmt.Println(len(service.Items)) // 0 - 空 map
```

### 2. 只收集命名注册

Map 自动注入**只收集命名注册的实例**，不包括默认注册：

```go
// 默认注册（不会被收集到 map 中）
defaultDB := &Database{Host: "default"}
container.MustRegisterInstance(defaultDB, di.Singleton)

// 命名注册（会被收集到 map 中）
primary := &Database{Host: "primary"}
container.MustRegisterInstanceNamed("primary", primary, di.Singleton)

// 自动收集时只包含命名注册
container.MustRegister(NewDatabaseManager, di.Singleton)
// manager.Databases 只包含 "primary"，不包含 defaultDB
```

### 3. 键类型必须是 string

Map 自动注入只支持 `map[string]T` 类型，其他键类型不支持：

```go
// ✅ 支持
func NewManager(dbs map[string]*Database) *Manager

// ❌ 不支持（会尝试正常解析，可能失败）
func NewManager(dbs map[int]*Database) *Manager
```

### 4. 值类型必须匹配

只有完全匹配值类型的命名实例才会被收集：

```go
// 注册 *Database
container.MustRegisterInstanceNamed("db1", &Database{}, di.Singleton)

// ✅ 匹配 - 会收集 db1
func NewManager(dbs map[string]*Database) *Manager

// ❌ 不匹配 - 不会收集 db1
func NewManager(dbs map[string]Database) *Manager
```

## 最佳实践

### 1. 明确意图

如果需要特定的 map 内容，直接注册 map：

```go
// ✅ 明确的 map 内容
settings := map[string]string{"key": "value"}
container.MustRegisterInstance(settings, di.Singleton)
```

如果需要收集所有命名实例，使用命名注册：

```go
// ✅ 收集所有数据库
container.MustRegisterInstanceNamed("db1", db1, di.Singleton)
container.MustRegisterInstanceNamed("db2", db2, di.Singleton)
// 构造函数会自动收集
```

### 2. 使用接口

对于多态场景，使用接口类型：

```go
type ICache interface {
    Get(key string) string
}

// 注册为接口类型
container.MustRegisterInstanceAsNamed("redis", redisCache, (*ICache)(nil), di.Singleton)
container.MustRegisterInstanceAsNamed("memory", memoryCache, (*ICache)(nil), di.Singleton)

// 构造函数接收接口 map
func NewCacheManager(caches map[string]ICache) *CacheManager {
    return &CacheManager{Caches: caches}
}
```

### 3. 文档化行为

在构造函数注释中说明 map 参数的来源：

```go
// NewDatabaseManager 创建数据库管理器
// dbs: 自动注入所有命名注册的 *Database 实例，以名称为键
func NewDatabaseManager(dbs map[string]*Database) *DatabaseManager {
    return &DatabaseManager{Databases: dbs}
}
```

### 4. 命名规范

使用清晰的命名约定：

```go
// ✅ 清晰的命名
container.MustRegisterInstanceNamed("primary", primaryDB, di.Singleton)
container.MustRegisterInstanceNamed("replica-1", replica1DB, di.Singleton)
container.MustRegisterInstanceNamed("replica-2", replica2DB, di.Singleton)

// ❌ 不清晰的命名
container.MustRegisterInstanceNamed("db1", primaryDB, di.Singleton)
container.MustRegisterInstanceNamed("db2", replica1DB, di.Singleton)
```

## 与切片自动注入的对比

| 特性 | Map 自动注入 | 切片自动注入 |\n|------|-------------|-------------|\n| 参数类型 | `map[string]T` | `[]T` |\n| 收集来源 | 只收集命名注册 | 收集默认注册 + 命名注册 |\n| 键/索引 | 使用注册名称作为键 | 无序索引 |\n| 访问方式 | 通过名称访问 | 通过索引或遍历 |\n| 适用场景 | 需要通过名称区分实例 | 需要处理所有实例 |\n\n```go
// Map 自动注入 - 通过名称访问
func NewManager(dbs map[string]*Database) *Manager {
    primary := dbs["primary"]  // 通过名称访问
    return &Manager{Primary: primary}
}

// 切片自动注入 - 遍历所有实例
func NewManager(dbs []*Database) *Manager {
    for _, db := range dbs {  // 遍历所有
        db.Connect()
    }
    return &Manager{Databases: dbs}
}
```

## 与 ResolveNamed 的区别

| 特性 | Map 自动注入 | ResolveNamed |\n|------|-------------|--------------|\n| 使用场景 | 构造函数依赖 | 手动解析单个实例 |\n| 调用方式 | 自动 | 手动调用 |\n| 返回结果 | 所有命名实例的 map | 单个命名实例 |\n| 适用范围 | 依赖注入 | 任何时候 |\n\n```go
// Map 自动注入（自动）
container.MustRegister(NewDatabaseManager, di.Singleton)
var manager *DatabaseManager
container.MustResolve(&manager)
// manager.Databases 包含所有命名实例

// ResolveNamed（手动）
var primaryDB *Database
container.MustResolveNamed("primary", &primaryDB)
// 只获取 "primary" 实例
```

## 完整示例代码

完整的示例代码请参考 `main.go` 中的示例 10-13：

- **示例 10**：多数据库连接（实际项目场景）
- **示例 11**：消息队列（实际项目场景）
- **示例 12**：缓存策略（实际项目场景）
- **示例 13**：微服务客户端（实际项目场景）

运行示例：
```bash
go run main.go
```

## 测试

完整的测试代码请参考 `di/map_injection_test.go`：

```bash
# 运行 map 自动注入测试
go test ./di -run TestMapAutoInjection -v
```

---

**版本**: v1.5.0
**更新日期**: 2026-02-02

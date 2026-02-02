package main

import (
	"fmt"
	"gofac/di"
)

// ==================== 示例1：RegisterInstance 基础用法 ====================

// Config 配置结构体
type Config struct {
	AppName string
	Port    int
}

func ExampleRegisterInstance() {
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
	// Output:
	// AppName: MyApp, Port: 8080
	// Same instance: true
}

// ==================== 示例2：RegisterInstanceAs 接口注册 ====================

// ILogger 日志接口
type ILogger interface {
	Log(msg string)
}

// ConsoleLogger 控制台日志实现
type ConsoleLogger struct {
	Prefix string
}

func (l *ConsoleLogger) Log(msg string) {
	fmt.Printf("[%s] %s\n", l.Prefix, msg)
}

func ExampleRegisterInstanceAs() {
	container := di.NewContainer()

	// 创建日志实例
	logger := &ConsoleLogger{Prefix: "INFO"}

	// 注册实例为接口类型
	container.MustRegisterInstanceAs(logger, (*ILogger)(nil), di.Singleton)

	// 通过接口类型解析
	var resolvedLogger ILogger
	container.MustResolve(&resolvedLogger)

	resolvedLogger.Log("Hello from instance registration!")
	// Output:
	// [INFO] Hello from instance registration!
}

// ==================== 示例3：Scoped 实例注册 ====================

type RequestContext struct {
	RequestID string
}

func ExampleRegisterInstanceScoped() {
	container := di.NewContainer()

	// 注册为Scoped（每个作用域独立）
	ctx1 := &RequestContext{RequestID: "req-001"}
	container.MustRegisterInstance(ctx1, di.Scoped)

	// 创建两个作用域
	scope1 := container.NewScope()
	scope2 := container.NewScope()

	// 从scope1解析
	var ctx1Resolved *RequestContext
	scope1.MustResolve(&ctx1Resolved)
	fmt.Printf("Scope1 RequestID: %s\n", ctx1Resolved.RequestID)

	// 从scope2解析（获取相同的实例，因为Scoped实例是共享的）
	var ctx2Resolved *RequestContext
	scope2.MustResolve(&ctx2Resolved)
	fmt.Printf("Scope2 RequestID: %s\n", ctx2Resolved.RequestID)

	// 两个作用域获取的是同一个实例
	fmt.Printf("Same instance: %v\n", ctx1Resolved == ctx2Resolved)
	// Output:
	// Scope1 RequestID: req-001
	// Scope2 RequestID: req-001
	// Same instance: true
}

// ==================== 示例4：切片类型支持 ====================

// UserService 用户服务（依赖切片类型）
type UserService struct {
	AllowedRoles []string
}

func NewUserService(roles []string) *UserService {
	return &UserService{AllowedRoles: roles}
}

func ExampleSliceType() {
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
	// Output:
	// Allowed roles: [admin user guest]
}

// ==================== 示例5：Map类型支持 ====================

type ISettings = map[string]string

// ConfigService 配置服务（依赖map类型）
type ConfigService struct {
	Settings ISettings
}

func NewConfigService(settings ISettings) *ConfigService {
	return &ConfigService{Settings: settings}
}

func ExampleMapType() {
	container := di.NewContainer()

	// 注册map实例
	settings := ISettings{
		"db_host": "localhost",
		"db_port": "5432",
		"db_name": "mydb",
	}
	container.MustRegisterInstance(settings, di.Singleton)

	// 注册依赖map的服务
	container.MustRegister(NewConfigService, di.Singleton)

	// 解析服务
	var configService *ConfigService
	container.MustResolve(&configService)

	fmt.Printf("DB Host: %s\n", configService.Settings["db_host"])
	fmt.Printf("DB Port: %s\n", configService.Settings["db_port"])
	// Output:
	// DB Host: localhost
	// DB Port: 5432
}

// ==================== 示例6：数组类型支持 ====================

type PriorityQueue struct {
	Priorities [5]int
}

func NewPriorityQueue(priorities [5]int) *PriorityQueue {
	return &PriorityQueue{Priorities: priorities}
}

func ExampleArrayType() {
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
	// Output:
	// Priorities: [1 2 3 4 5]
}

// ==================== 示例7：全局容器便捷方法 ====================

func ExampleGlobalRegisterInstance() {
	// 重置全局容器
	di.GlobalReset()

	// 使用全局容器注册实例
	config := &Config{AppName: "GlobalApp", Port: 9000}
	di.MustRegisterInstance(config, di.Singleton)

	// 使用泛型方法解析
	resolvedConfig := di.MustGet[*Config]()

	fmt.Printf("Global AppName: %s, Port: %d\n", resolvedConfig.AppName, resolvedConfig.Port)
	// Output:
	// Global AppName: GlobalApp, Port: 9000
}

// ==================== 示例8：复杂引用类型组合 ====================

type DatabaseConfig struct {
	Hosts    []string
	Ports    map[string]int
	Replicas [3]string
}

type Database struct {
	Config       *DatabaseConfig
	Settings     map[string]string
	AllowedRoles []string
	UserService  *UserService
}

func NewDatabase(config *DatabaseConfig, settings map[string]string, allowedRoles []string, userService *UserService) *Database {
	return &Database{Config: config, Settings: settings, AllowedRoles: allowedRoles, UserService: userService}
}

func ExampleComplexReferenceTypes() {
	container := di.NewContainer()

	// 注册复杂配置实例
	dbConfig := &DatabaseConfig{
		Hosts:    []string{"host1", "host2", "host3"},
		Ports:    map[string]int{"primary": 5432, "replica": 5433},
		Replicas: [3]string{"replica1", "replica2", "replica3"},
	}
	settings := map[string]string{
		"db_host": "localhost",
		"db_port": "5432",
		"db_name": "mydb",
	}
	allowedRoles := []string{"admin", "user", "guest"}
	userService := NewUserService(allowedRoles)

	container.MustRegisterInstance(dbConfig, di.Singleton)
	container.MustRegisterInstance(settings, di.Singleton)
	container.MustRegisterInstance(allowedRoles, di.Singleton)
	container.MustRegisterInstanceAs(userService, (*UserService)(nil), di.Singleton)
	//container.MustRegisterInstance(userService, di.Singleton)

	// 注册数据库服务
	container.MustRegister(NewDatabase, di.Singleton)

	// 解析服务
	var db *Database
	container.MustResolve(&db)

	fmt.Printf("Hosts: %v\n", db.Config.Hosts)
	fmt.Printf("Primary Port: %d\n", db.Config.Ports["primary"])
	fmt.Printf("Replicas: %v\n", db.Config.Replicas)
	fmt.Printf("Allowed Roles: %v\n", db.AllowedRoles)
	fmt.Printf("DB Host: %s\n", db.Settings["db_host"])
	fmt.Printf("DB Port: %s\n", db.Settings["db_port"])
	fmt.Printf("Allowed Roles: %s\n", db.UserService.AllowedRoles)

	// Output:
	// Hosts: [host1 host2 host3]
	// Primary Port: 5432
	// Replicas: [replica1 replica2 replica3]
}

type ArrayUserService struct {
	Services []*UserService
}

func NewArrayUserService(services []*UserService) *ArrayUserService {
	return &ArrayUserService{Services: services}
}

type ArrayUserService2 struct {
	Services *ArrayUserService
}

func NewArrayUserService2(services *ArrayUserService) *ArrayUserService2 {
	return &ArrayUserService2{Services: services}
}

type MapUserService struct {
	ServiceMap map[string]*UserService
}

func NewMapUserService(serviceMap map[string]*UserService) *MapUserService {
	return &MapUserService{ServiceMap: serviceMap}
}

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

	// 解析所有同类型的服务
	var services []*UserService
	container.MustResolveAll(&services)

	fmt.Printf("Total services: %d\n", len(services))
	for i, service := range services {
		fmt.Printf("Service %d - Allowed Roles: %v\n", i+1, service.AllowedRoles)
	}

	// 手动创建 ArrayUserService 并注册
	//arrayUserService := NewArrayUserService(services)
	container.MustRegister(NewArrayUserService, di.Singleton)

	// 注册 ArrayUserService2（会自动注入 *ArrayUserService 依赖）
	container.MustRegister(NewArrayUserService2, di.Singleton)

	// 解析嵌套结构
	var arrayUserService2 *ArrayUserService2
	container.MustResolve(&arrayUserService2)
	fmt.Printf("\nNested structure - First service roles: %v\n", arrayUserService2.Services.Services[0].AllowedRoles)
	fmt.Printf("Nested structure - Total services in array: %d\n", len(arrayUserService2.Services.Services))

	container.MustRegister(NewMapUserService, di.Singleton)
	var mapUserService *MapUserService
	container.MustResolve(&mapUserService)
	fmt.Printf("\nMap structure - Total services in map: %d\n", len(mapUserService.ServiceMap))
}

// ==================== 示例10：多数据库连接（实际项目场景）====================

type DBConnection struct {
	Name string
	Host string
	Port int
}

func ExampleMultipleDatabases() {
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
	for _, db := range allDBs {
		fmt.Printf("  - %s: %s:%d\n", db.Name, db.Host, db.Port)
	}
}

// ==================== 示例11：消息队列（实际项目场景）====================

type MessageQueue struct {
	Name  string
	Topic string
}

func (mq *MessageQueue) Publish(message string) {
	fmt.Printf("[%s] Publishing to %s: %s\n", mq.Name, mq.Topic, message)
}

func ExampleMessageQueues() {
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
	fmt.Printf("\nBroadcasting to all %d queues:\n", len(allQueues))
	for _, queue := range allQueues {
		queue.Publish("System maintenance at 2AM")
	}
}

// ==================== 示例12：缓存策略（实际项目场景）====================

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
	fmt.Printf("[Redis:%s] Set %s = %s\n", r.Name, key, value)
}

type MemoryCache struct {
	Name string
}

func (m *MemoryCache) Get(key string) string {
	return fmt.Sprintf("[Memory:%s] %s", m.Name, key)
}

func (m *MemoryCache) Set(key string, value string) {
	fmt.Printf("[Memory:%s] Set %s = %s\n", m.Name, key, value)
}

func ExampleCacheStrategies() {
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
	fmt.Printf("\nAll cache implementations (%d):\n", len(allCaches))
	for _, cache := range allCaches {
		fmt.Println(cache.Get("test-key"))
	}
}

// ==================== 示例13：微服务客户端（实际项目场景）====================

type ServiceClient struct {
	ServiceName string
	BaseURL     string
}

func (sc *ServiceClient) Call(endpoint string) string {
	return fmt.Sprintf("[%s] Calling %s%s", sc.ServiceName, sc.BaseURL, endpoint)
}

func ExampleMicroserviceClients() {
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
	fmt.Printf("\nHealth check for %d services:\n", len(allClients))
	for _, client := range allClients {
		fmt.Println(client.Call("/health"))
	}
}

// ==================== 主函数：运行所有示例 ====================

func main() {
	fmt.Println("========== 示例1：RegisterInstance 基础用法 ==========")
	ExampleRegisterInstance()

	fmt.Println("\n========== 示例2：RegisterInstanceAs 接口注册 ==========")
	ExampleRegisterInstanceAs()

	fmt.Println("\n========== 示例3：Scoped 实例注册 ==========")
	ExampleRegisterInstanceScoped()

	fmt.Println("\n========== 示例4：切片类型支持 ==========")
	ExampleSliceType()

	fmt.Println("\n========== 示例5：Map类型支持 ==========")
	ExampleMapType()

	fmt.Println("\n========== 示例6：数组类型支持 ==========")
	ExampleArrayType()

	fmt.Println("\n========== 示例7：全局容器便捷方法 ==========")
	ExampleGlobalRegisterInstance()

	fmt.Println("\n========== 示例8：复杂引用类型组合 ==========")
	ExampleComplexReferenceTypes()

	fmt.Println("\n========== 示例9：命名注册 - 多个同类型实例 ==========")
	ExampleMoreComplexReferenceTypes()

	fmt.Println("\n========== 示例10：多数据库连接（实际项目场景）==========")
	ExampleMultipleDatabases()

	fmt.Println("\n========== 示例11：消息队列（实际项目场景）==========")
	ExampleMessageQueues()

	fmt.Println("\n========== 示例12：缓存策略（实际项目场景）==========")
	ExampleCacheStrategies()

	fmt.Println("\n========== 示例13：微服务客户端（实际项目场景）==========")
	ExampleMicroserviceClients()
}

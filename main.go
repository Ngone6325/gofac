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
	Config *DatabaseConfig
}

func NewDatabase(config *DatabaseConfig) *Database {
	return &Database{Config: config}
}

func ExampleComplexReferenceTypes() {
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
	// Output:
	// Hosts: [host1 host2 host3]
	// Primary Port: 5432
	// Replicas: [replica1 replica2 replica3]
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
}

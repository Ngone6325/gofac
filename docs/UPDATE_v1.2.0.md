# Gofac v1.2.0 更新说明

## 更新概述

本次更新为 `RegisterAs` 和 `RegisterInstanceAs` 方法添加了**具体类型注册**支持，使其能够同时支持接口类型和具体类型作为目标类型。

## 新增功能

### RegisterAs 和 RegisterInstanceAs 支持具体类型

之前这两个方法只支持接口类型注册，现在可以注册为任意具体类型。

#### 使用示例

**之前（仅支持接口）：**
```go
// ✅ 支持：注册为接口
container.RegisterAs(NewLogger, (*ILogger)(nil), di.Singleton)

// ❌ 不支持：注册为具体类型
container.RegisterAs(NewService, (*UserService)(nil), di.Singleton)
```

**现在（同时支持接口和具体类型）：**
```go
// ✅ 支持：注册为接口
container.RegisterAs(NewLogger, (*ILogger)(nil), di.Singleton)

// ✅ 支持：注册为具体类型
container.RegisterAs(NewService, (*UserService)(nil), di.Singleton)
```

## 使用场景

### 1. 精确控制服务类型

```go
type UserService struct {
    Name string
}

func NewUserService() *UserService {
    return &UserService{Name: "service"}
}

// 注册为具体类型 *UserService
container.MustRegisterAs(NewUserService, (*UserService)(nil), di.Singleton)

// 其他服务可以依赖 *UserService
type OrderService struct {
    UserService *UserService
}

func NewOrderService(us *UserService) *OrderService {
    return &OrderService{UserService: us}
}

container.MustRegister(NewOrderService, di.Singleton)
```

### 2. 同一实现注册为多个类型

```go
type Service struct {
    Value string
}

func (s *Service) GetValue() string {
    return s.Value
}

type IService interface {
    GetValue() string
}

// 注册为接口类型
container.MustRegisterAs(NewService1, (*IService)(nil), di.Singleton)

// 同时注册为具体类型（不同实例）
container.MustRegisterAs(NewService2, (*Service)(nil), di.Singleton)

// 可以通过两种方式解析
iface := di.MustGet[IService]()      // 获取接口实例
concrete := di.MustGet[*Service]()   // 获取具体类型实例
```

### 3. 测试中的类型替换

```go
// 生产代码
type RealService struct {
    DB *Database
}

// 测试代码
type MockService struct {
    Data map[string]string
}

// 在测试中，注册 MockService 为 RealService 类型
mockService := &MockService{Data: map[string]string{"key": "value"}}
container.MustRegisterInstanceAs(mockService, (*RealService)(nil), di.Singleton)

// 依赖 *RealService 的代码无需修改
```

## 技术实现

### 修改的代码

**文件：`di/container.go`**

修改了 `register` 和 `registerInstance` 方法中的类型解析逻辑：

```go
// 判断是指向接口还是具体类型
if elemType.Kind() == reflect.Interface {
    // 接口类型：使用接口类型作为服务类型
    svcType = elemType
    if !implType.Implements(svcType) {
        return fmt.Errorf("类型%s未实现接口%s", implType, svcType)
    }
} else {
    // 具体类型：使用完整的指针类型作为服务类型
    // 例如：(*UserService)(nil) -> 注册为 *UserService 类型
    svcType = targetType
    if !isTypeCompatible(implType, svcType) {
        return fmt.Errorf("类型%s无法转换为目标类型%s", implType, svcType)
    }
}
```

**关键变化：**
- 之前：具体类型使用 `elemType`（值类型）
- 现在：具体类型使用 `targetType`（指针类型）

这确保了 `(*UserService)(nil)` 注册为 `*UserService` 而不是 `UserService`。

## 测试覆盖

新增测试文件：`di/concrete_type_test.go`

包含 9 个测试用例：

1. ✅ `TestRegisterAs_ConcreteType` - 构造函数注册为具体类型
2. ✅ `TestRegisterAs_ConcreteType_AsDependency` - 具体类型作为依赖
3. ✅ `TestRegisterInstanceAs_ConcreteType` - 实例注册为具体类型
4. ✅ `TestRegisterInstanceAs_ConcreteType_AsDependency` - 具体类型实例作为依赖
5. ✅ `TestRegisterAs_InterfaceAndConcreteType` - 同时注册接口和具体类型
6. ✅ `TestRegisterInstanceAs_InterfaceAndConcreteType` - 实例同时注册为接口和具体类型
7. ✅ `TestComplexDependencyGraph_WithConcreteTypes` - 复杂依赖图
8. ✅ `TestMustGet_WithConcreteType` - 泛型方法解析具体类型
9. ✅ `TestMustRegisterInstanceAs_WithConcreteType_AndGenericGet` - 实例注册 + 泛型解析

所有测试通过：
```
PASS
ok      gofac/di        0.008s
```

## 文档更新

### 新增文档

- **CONCRETE_TYPE_SUPPORT.md** - 具体类型支持的详细文档
  - 使用方式
  - 使用场景
  - 类型兼容性
  - 语法说明
  - 完整示例
  - 注意事项

### 更新文档

- **README.md** - 更新特性列表和注册方式说明

## 向后兼容性

✅ **完全向后兼容**

- 所有原有代码无需修改
- 接口注册功能保持不变
- 新功能为可选特性

## API 变化

### 无破坏性变更

所有 API 签名保持不变：

```go
// 方法签名未变
func (c *Container) RegisterAs(ctor any, interfaceType any, scope LifetimeScope) error
func (c *Container) RegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) error
```

### 行为增强

- `interfaceType` 参数现在接受：
  - `(*IInterface)(nil)` - 接口类型（原有功能）
  - `(*StructType)(nil)` - 具体类型（新功能）

## 使用建议

### 何时使用具体类型注册

**推荐场景：**
- ✅ 需要精确控制服务类型
- ✅ 同一实现需要注册为多个类型
- ✅ 测试中的类型替换
- ✅ 复杂依赖图中的类型管理

**不推荐场景：**
- ❌ 简单的服务注册（使用 `Register` 即可）
- ❌ 只需要接口抽象（使用接口注册）

### 接口 vs 具体类型

| 场景 | 推荐方式 |
|------|---------|
| 需要抽象和解耦 | 接口注册 |
| 需要精确类型控制 | 具体类型注册 |
| 测试替换 | 具体类型注册 |
| 简单服务 | 直接使用 `Register` |

## 示例代码

完整示例请参考：
- `main.go` - 示例 8：复杂引用类型组合
- `di/concrete_type_test.go` - 完整测试用例

## 版本信息

- **版本号**: v1.2.0
- **发布日期**: 2026-02-02
- **兼容性**: 完全向后兼容 v1.1.0

## 总结

本次更新通过简单的逻辑修改，为 `RegisterAs` 和 `RegisterInstanceAs` 方法添加了强大的具体类型注册功能，使 gofac 的类型系统更加灵活和强大。

**主要改进：**
- ✅ 支持具体类型注册
- ✅ 完全向后兼容
- ✅ 完整的测试覆盖
- ✅ 详细的文档说明
- ✅ 无破坏性变更

---

**作者**: Gofac Team
**更新时间**: 2026-02-02

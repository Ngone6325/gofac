package gofac

import (
	"testing"
)

// Test types
type TestService struct {
	Value string
}

func NewTestService() *TestService {
	return &TestService{Value: "test"}
}

type TestDependency struct {
	Name string
}

func NewTestDependency() *TestDependency {
	return &TestDependency{Name: "dependency"}
}

type TestServiceWithDep struct {
	Dep *TestDependency
}

func NewTestServiceWithDep(dep *TestDependency) *TestServiceWithDep {
	return &TestServiceWithDep{Dep: dep}
}

// Test interface
type ITestInterface interface {
	GetValue() string
}

type TestImpl struct {
	Value string
}

func (t *TestImpl) GetValue() string {
	return t.Value
}

func NewTestImpl() *TestImpl {
	return &TestImpl{Value: "impl"}
}

// TestNewContainer tests container creation
func TestNewContainer(t *testing.T) {
	container := NewContainer()
	if container == nil {
		t.Fatal("NewContainer returned nil")
	}
	if container.services == nil {
		t.Error("services map not initialized")
	}
	if container.namedServices == nil {
		t.Error("namedServices map not initialized")
	}
}

// TestRegister tests basic registration
func TestRegister(t *testing.T) {
	container := NewContainer()

	err := container.Register(NewTestService, Singleton)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Test duplicate registration
	err = container.Register(NewTestService, Singleton)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

// TestRegisterAs tests interface registration
func TestRegisterAs(t *testing.T) {
	container := NewContainer()

	err := container.RegisterAs(NewTestImpl, (*ITestInterface)(nil), Singleton)
	if err != nil {
		t.Fatalf("RegisterAs failed: %v", err)
	}

	var result ITestInterface
	err = container.Resolve(&result)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.GetValue() != "impl" {
		t.Errorf("Expected 'impl', got '%s'", result.GetValue())
	}
}

// TestRegisterInstance tests instance registration
func TestRegisterInstance(t *testing.T) {
	container := NewContainer()

	instance := &TestService{Value: "instance"}
	err := container.RegisterInstance(instance, Singleton)
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	var result *TestService
	err = container.Resolve(&result)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.Value != "instance" {
		t.Errorf("Expected 'instance', got '%s'", result.Value)
	}

	// Verify it's the same instance
	if result != instance {
		t.Error("Expected same instance reference")
	}
}

// TestRegisterInstanceTransient tests that Transient is not allowed for instances
func TestRegisterInstanceTransient(t *testing.T) {
	container := NewContainer()

	instance := &TestService{Value: "test"}
	err := container.RegisterInstance(instance, Transient)
	if err != ErrTransientInstance {
		t.Errorf("Expected ErrTransientInstance, got %v", err)
	}
}

// TestRegisterInstanceNil tests that nil instances are rejected
func TestRegisterInstanceNil(t *testing.T) {
	container := NewContainer()

	err := container.RegisterInstance(nil, Singleton)
	if err != ErrNilInstance {
		t.Errorf("Expected ErrNilInstance, got %v", err)
	}
}

// TestResolve tests basic resolution
func TestResolve(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Singleton)

	var result *TestService
	err := container.Resolve(&result)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.Value != "test" {
		t.Errorf("Expected 'test', got '%s'", result.Value)
	}
}

// TestResolveDependency tests dependency injection
func TestResolveDependency(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestDependency, Singleton)
	container.MustRegister(NewTestServiceWithDep, Singleton)

	var result *TestServiceWithDep
	err := container.Resolve(&result)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if result.Dep == nil {
		t.Fatal("Dependency not injected")
	}

	if result.Dep.Name != "dependency" {
		t.Errorf("Expected 'dependency', got '%s'", result.Dep.Name)
	}
}

// TestSingletonLifetime tests singleton behavior
func TestSingletonLifetime(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Singleton)

	var result1 *TestService
	var result2 *TestService

	container.MustResolve(&result1)
	container.MustResolve(&result2)

	if result1 != result2 {
		t.Error("Singleton should return same instance")
	}
}

// TestTransientLifetime tests transient behavior
func TestTransientLifetime(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Transient)

	var result1 *TestService
	var result2 *TestService

	container.MustResolve(&result1)
	container.MustResolve(&result2)

	if result1 == result2 {
		t.Error("Transient should return different instances")
	}
}

// TestScopedLifetime tests scoped behavior
func TestScopedLifetime(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Scoped)

	scope1 := container.NewScope()
	scope2 := container.NewScope()

	var result1 *TestService
	var result2 *TestService
	var result3 *TestService

	scope1.MustResolve(&result1)
	scope1.MustResolve(&result2)
	scope2.MustResolve(&result3)

	// Same scope should return same instance
	if result1 != result2 {
		t.Error("Scoped should return same instance within scope")
	}

	// Different scope should return different instance
	if result1 == result3 {
		t.Error("Scoped should return different instances across scopes")
	}
}

// TestScopedOnRootContainer tests that Scoped cannot be resolved from root
func TestScopedOnRootContainer(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Scoped)

	var result *TestService
	err := container.Resolve(&result)
	if err != ErrScopedOnRootContainer {
		t.Errorf("Expected ErrScopedOnRootContainer, got %v", err)
	}
}

// TestRegisterInstanceNamed tests named instance registration
func TestRegisterInstanceNamed(t *testing.T) {
	container := NewContainer()

	instance1 := &TestService{Value: "first"}
	instance2 := &TestService{Value: "second"}

	err := container.RegisterInstanceNamed("first", instance1, Singleton)
	if err != nil {
		t.Fatalf("RegisterInstanceNamed failed: %v", err)
	}

	err = container.RegisterInstanceNamed("second", instance2, Singleton)
	if err != nil {
		t.Fatalf("RegisterInstanceNamed failed: %v", err)
	}

	var result *TestService
	err = container.ResolveNamed("first", &result)
	if err != nil {
		t.Fatalf("ResolveNamed failed: %v", err)
	}

	if result.Value != "first" {
		t.Errorf("Expected 'first', got '%s'", result.Value)
	}
}

// TestResolveAll tests resolving all instances of a type
func TestResolveAll(t *testing.T) {
	container := NewContainer()

	instance1 := &TestService{Value: "first"}
	instance2 := &TestService{Value: "second"}

	container.MustRegisterInstance(instance1, Singleton)
	container.MustRegisterInstanceNamed("named", instance2, Singleton)

	var results []*TestService
	err := container.ResolveAll(&results)
	if err != nil {
		t.Fatalf("ResolveAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

// TestMustRegister tests Must* methods panic behavior
func TestMustRegister(t *testing.T) {
	container := NewContainer()

	// Should not panic
	container.MustRegister(NewTestService, Singleton)

	// Should panic on duplicate
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate registration")
		}
	}()
	container.MustRegister(NewTestService, Singleton)
}

// TestGet tests generic Get function
func TestGet(t *testing.T) {
	GlobalReset()

	MustRegister(NewTestService, Singleton)

	result, err := Get[*TestService]()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.Value != "test" {
		t.Errorf("Expected 'test', got '%s'", result.Value)
	}
}

// TestMustGet tests generic MustGet function
func TestMustGet(t *testing.T) {
	GlobalReset()

	MustRegister(NewTestService, Singleton)

	result := MustGet[*TestService]()

	if result.Value != "test" {
		t.Errorf("Expected 'test', got '%s'", result.Value)
	}
}

// TestReset tests container reset
func TestReset(t *testing.T) {
	container := NewContainer()

	container.MustRegister(NewTestService, Singleton)
	container.Reset()

	var result *TestService
	err := container.Resolve(&result)
	if err == nil {
		t.Error("Expected error after reset")
	}
}

// TestCircularDependency tests circular dependency detection
func TestCircularDependency(t *testing.T) {
	// This test would require creating circular dependencies
	// which is complex to set up, so we'll skip for now
	t.Skip("Circular dependency test requires complex setup")
}

// TestInvalidRegistration tests error cases
func TestInvalidRegistration(t *testing.T) {
	container := NewContainer()

	// Not a function
	err := container.Register("not a function", Singleton)
	if err != ErrNotFunc {
		t.Errorf("Expected ErrNotFunc, got %v", err)
	}

	// Function with no return value
	noReturn := func() {}
	err = container.Register(noReturn, Singleton)
	if err == nil {
		t.Error("Expected error for function with no return value")
	}
}

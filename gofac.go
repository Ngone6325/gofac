package gofac

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceDef Service definition: stores registration metadata, cached parameter types, and singleton instances
type ServiceDef struct {
	implType   reflect.Type   // Service implementation type (constructor return value or instance type)
	scope      LifetimeScope  // Lifetime scope
	instance   reflect.Value  // Singleton instance cache or pre-registered instance
	ctor       reflect.Value  // Constructor reflection value (empty for instance registration)
	ctorType   reflect.Type   // Constructor reflection type (empty for instance registration)
	once       sync.Once      // Atomic operation for singleton instance initialization
	paramTypes []reflect.Type // Cached constructor parameter types (core optimization)
	paramOnce  sync.Once      // Ensures parameter types are parsed only once (concurrency-safe)
	isInstance bool           // Whether this is an instance registration (if true, use instance directly without calling ctor)
}

// Container DI container core: manages all services with concurrency safety
type Container struct {
	services      map[reflect.Type]*ServiceDef            // Default (unnamed) services
	namedServices map[string]map[reflect.Type]*ServiceDef // Named services: name -> type -> ServiceDef
	mu            sync.RWMutex
}

// Scope Within the same Scope, Scoped instances are unique; different Scopes are isolated from each other
type Scope struct {
	root       *Container                     // Associated root container (shares registration metadata)
	scopedInst map[reflect.Type]reflect.Value // Scoped instance cache for this scope
	mu         sync.RWMutex                   // Scope concurrency-safe lock
}

// NewContainer Creates a new DI container
func NewContainer() *Container {
	return &Container{
		services:      make(map[reflect.Type]*ServiceDef),
		namedServices: make(map[string]map[reflect.Type]*ServiceDef),
	}
}

// Global container: for single-service architecture, eliminates manual container creation
var Global = NewContainer()

// Register Basic registration: registers by constructor return value type, returns error (requires manual handling)
func (c *Container) Register(ctor any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.register(ctor, nil, scope)
}

// RegisterAs Interface registration: registers implementation type as specified interface type, returns error (requires manual handling)
func (c *Container) RegisterAs(ctor any, interfaceType any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.register(ctor, interfaceType, scope)
}

// register Internal common registration logic, extracts duplicate code
func (c *Container) register(ctor any, interfaceType any, scope LifetimeScope) error {
	// Parse constructor reflection information
	ctorVal := reflect.ValueOf(ctor)
	ctorType := ctorVal.Type()
	if ctorType.Kind() != reflect.Func {
		return ErrNotFunc
	}

	// Validate constructor return value: only 1 return value, and must be concrete type
	numOut := ctorType.NumOut()
	if numOut != 1 {
		return fmt.Errorf("%w, current return value count: %d", ErrNoReturn, numOut)
	}
	implType := ctorType.Out(0)
	if implType.Kind() == reflect.Interface {
		return fmt.Errorf("%w, return value is interface: %s", ErrNotConcreteType, implType)
	}

	// Determine final registered service type (interface/implementation type)
	svcType := implType
	if interfaceType != nil {
		// Parse target type
		targetType := reflect.TypeOf(interfaceType)

		// Check if it's a pointer type
		if targetType.Kind() != reflect.Ptr {
			return ErrInvalidInterfaceType
		}

		// Get the element type pointed to by the pointer
		elemType := targetType.Elem()

		// Determine if it points to an interface or concrete type
		if elemType.Kind() == reflect.Interface {
			// Interface type: use interface type as service type
			svcType = elemType
			if !implType.Implements(svcType) {
				return fmt.Errorf("type %s does not implement interface %s", implType, svcType)
			}
		} else {
			// Concrete type: use complete pointer type as service type
			// Example: (*UserService)(nil) -> register as *UserService type
			svcType = targetType
			// Enhanced type compatibility check, supports pointer/value type conversion
			if !isTypeCompatible(implType, svcType) {
				return fmt.Errorf("type %s cannot be converted to target type %s", implType, svcType)
			}
		}
	}

	// Check for duplicate registration
	if _, exists := c.services[svcType]; exists {
		return fmt.Errorf("%w, type: %s", ErrRegisterDuplicate, svcType)
	}

	// Encapsulate service definition and add to container
	c.services[svcType] = &ServiceDef{
		implType:   implType,
		scope:      scope,
		ctor:       ctorVal,
		ctorType:   ctorType,
		isInstance: false,
	}
	return nil
}

// RegisterInstance Instance registration: directly registers a created instance, registers by instance type
// Note: Does not support Transient lifetime (instance already created, cannot return new instance each time)
func (c *Container) RegisterInstance(instance any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstance(instance, nil, scope)
}

// RegisterInstanceAs Instance interface registration: registers a created instance as specified interface type
// Note: Does not support Transient lifetime (instance already created, cannot return new instance each time)
func (c *Container) RegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstance(instance, interfaceType, scope)
}

// registerInstance Internal instance registration logic
func (c *Container) registerInstance(instance any, interfaceType any, scope LifetimeScope) error {
	// Transient does not support instance registration (cannot create new instance each time)
	if scope == Transient {
		return ErrTransientInstance
	}

	// Validate instance is not nil
	if instance == nil {
		return ErrNilInstance
	}

	instVal := reflect.ValueOf(instance)
	implType := instVal.Type()

	// Determine final registered service type (interface/implementation type)
	svcType := implType
	if interfaceType != nil {
		// Parse target type
		targetType := reflect.TypeOf(interfaceType)

		// Check if it's a pointer type
		if targetType.Kind() != reflect.Ptr {
			return ErrInvalidInterfaceType
		}

		// Get the element type pointed to by the pointer
		elemType := targetType.Elem()

		// Determine if it points to an interface or concrete type
		if elemType.Kind() == reflect.Interface {
			// Interface type: use interface type as service type
			svcType = elemType
			if !implType.Implements(svcType) {
				return fmt.Errorf("instance type %s does not implement interface %s", implType, svcType)
			}
		} else {
			// Concrete type: use complete pointer type as service type
			// Example: (*UserService)(nil) -> register as *UserService type
			svcType = targetType
			// Enhanced type compatibility check, supports pointer/value type conversion
			if !isTypeCompatible(implType, svcType) {
				return fmt.Errorf("instance type %s cannot be converted to target type %s", implType, svcType)
			}
		}
	}

	// Check for duplicate registration
	if _, exists := c.services[svcType]; exists {
		return fmt.Errorf("%w, type: %s", ErrRegisterDuplicate, svcType)
	}

	// Encapsulate service definition and add to container
	c.services[svcType] = &ServiceDef{
		implType:   implType,
		scope:      scope,
		instance:   instVal,
		isInstance: true,
	}
	return nil
}

// RegisterInstanceNamed Named instance registration: registers an instance with a name, allows multiple instances of the same type
func (c *Container) RegisterInstanceNamed(name string, instance any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstanceNamed(name, instance, nil, scope)
}

// RegisterInstanceAsNamed Named instance interface registration: registers an instance with a name as specified type
func (c *Container) RegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.registerInstanceNamed(name, instance, interfaceType, scope)
}

// registerInstanceNamed Internal named instance registration logic
func (c *Container) registerInstanceNamed(name string, instance any, interfaceType any, scope LifetimeScope) error {
	// Transient does not support instance registration
	if scope == Transient {
		return ErrTransientInstance
	}

	// Validate instance is not nil
	if instance == nil {
		return ErrNilInstance
	}

	// Validate name is not empty
	if name == "" {
		return fmt.Errorf("name cannot be empty for named registration")
	}

	instVal := reflect.ValueOf(instance)
	implType := instVal.Type()

	// Determine final registered service type
	svcType := implType
	if interfaceType != nil {
		targetType := reflect.TypeOf(interfaceType)
		if targetType.Kind() != reflect.Ptr {
			return ErrInvalidInterfaceType
		}

		elemType := targetType.Elem()
		if elemType.Kind() == reflect.Interface {
			svcType = elemType
			if !implType.Implements(svcType) {
				return fmt.Errorf("instance type %s does not implement interface %s", implType, svcType)
			}
		} else {
			svcType = targetType
			if !isTypeCompatible(implType, svcType) {
				return fmt.Errorf("instance type %s cannot be converted to target type %s", implType, svcType)
			}
		}
	}

	// Initialize named services map
	if c.namedServices[name] == nil {
		c.namedServices[name] = make(map[reflect.Type]*ServiceDef)
	}

	// Check for duplicate registration
	if _, exists := c.namedServices[name][svcType]; exists {
		return fmt.Errorf("%w, name: %s, type: %s", ErrRegisterDuplicate, name, svcType)
	}

	// Encapsulate service definition and add to container
	c.namedServices[name][svcType] = &ServiceDef{
		implType:   implType,
		scope:      scope,
		instance:   instVal,
		isInstance: true,
	}
	return nil
}

// isTypeCompatible Checks if two types are compatible (supports pointer/value type conversion)
func isTypeCompatible(implType, targetType reflect.Type) bool {
	// Directly assignable (including same type)
	if implType.AssignableTo(targetType) {
		return true
	}

	// Convertible
	if implType.ConvertibleTo(targetType) {
		return true
	}

	// Check pointer type compatibility: if implementation is value type, target is corresponding pointer type
	if implType.Kind() != reflect.Ptr && reflect.PointerTo(implType).AssignableTo(targetType) {
		return true
	}

	// Check reverse pointer type compatibility: if implementation is pointer type, target is corresponding value type
	if implType.Kind() == reflect.Ptr && implType.Elem().AssignableTo(targetType) {
		return true
	}

	return false
}

// Resolve Original resolution: receives instance through pointer, returns error (compatible with old logic)
func (c *Container) Resolve(out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}
	svcType := outVal.Elem().Type()
	instance, err := c.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return err
	}
	outVal.Elem().Set(instance)
	return nil
}

// ResolveNamed Named resolution: resolves specific service instance by name
func (c *Container) ResolveNamed(name string, out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}
	svcType := outVal.Elem().Type()

	c.mu.RLock()
	namedMap, exists := c.namedServices[name]
	if !exists {
		c.mu.RUnlock()
		return fmt.Errorf("named service does not exist, name: %s", name)
	}
	serviceDef, exists := namedMap[svcType]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w, name: %s, type: %s", ErrServiceNotRegistered, name, svcType)
	}

	// Named services currently only support instance registration, return instance directly
	if serviceDef.isInstance {
		outVal.Elem().Set(serviceDef.instance)
		return nil
	}

	return fmt.Errorf("named services do not support constructor registration yet, name: %s", name)
}

// ResolveAll Resolves all services of the same type (including default and all named services)
func (c *Container) ResolveAll(out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}

	// Check output type must be a slice pointer
	elemType := outVal.Elem().Type()
	if elemType.Kind() != reflect.Slice {
		return fmt.Errorf("ResolveAll output parameter must be a slice pointer, current type: %s", elemType)
	}

	// Get slice element type
	itemType := elemType.Elem()

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create result slice
	results := reflect.MakeSlice(elemType, 0, 0)

	// Add default service (if exists)
	if serviceDef, exists := c.services[itemType]; exists {
		if serviceDef.isInstance {
			results = reflect.Append(results, serviceDef.instance)
		}
	}

	// Add all named services
	for _, namedMap := range c.namedServices {
		if serviceDef, exists := namedMap[itemType]; exists {
			if serviceDef.isInstance {
				results = reflect.Append(results, serviceDef.instance)
			}
		}
	}

	// Set result
	outVal.Elem().Set(results)
	return nil
}

// resolve Internal recursive resolution core method: handles dependencies, caching, lifetime (original logic with added Scoped validation)
func (c *Container) resolve(svcType reflect.Type, track map[reflect.Type]bool) (reflect.Value, error) {
	// Read lock to get service definition, avoid write blocking
	c.mu.RLock()
	serviceDef, exists := c.services[svcType]
	c.mu.RUnlock()
	if !exists {
		return reflect.Value{}, fmt.Errorf("%w, type: %s", ErrServiceNotRegistered, svcType)
	}

	// Circular dependency detection
	if track[svcType] {
		return reflect.Value{}, fmt.Errorf("%w, circular dependency chain contains: %s", ErrResolveCircularDependency, svcType)
	}
	track[svcType] = true
	defer delete(track, svcType)

	// New: Scoped prohibits direct resolution from root container, must use scope
	if serviceDef.scope == Scoped {
		return reflect.Value{}, ErrScopedOnRootContainer
	}

	// Instance registration: directly return pre-registered instance (Singleton/Scoped)
	if serviceDef.isInstance {
		return serviceDef.instance, nil
	}

	// Singleton: return existing instance directly
	if serviceDef.scope == Singleton && serviceDef.instance.IsValid() {
		return serviceDef.instance, nil
	}

	// Core optimization: cache constructor parameter types, parse only on first resolution
	serviceDef.paramOnce.Do(func() {
		numIn := serviceDef.ctorType.NumIn()
		params := make([]reflect.Type, numIn)
		for i := 0; i < numIn; i++ {
			params[i] = serviceDef.ctorType.In(i)
		}
		serviceDef.paramTypes = params
	})
	paramTypes := serviceDef.paramTypes

	// Recursively resolve all dependency parameters
	params := make([]reflect.Value, len(paramTypes))
	for i, pType := range paramTypes {
		// Check if parameter is a slice type
		if pType.Kind() == reflect.Slice {
			// First try to resolve slice type directly (if registered)
			c.mu.RLock()
			_, sliceExists := c.services[pType]
			c.mu.RUnlock()

			if sliceExists {
				// Slice type is registered, resolve directly
				pInstance, err := c.resolve(pType, track)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
				}
				params[i] = pInstance
			} else {
				// Slice type not registered: automatically collect all instances of that element type
				elemType := pType.Elem()

				// Create result slice
				results := reflect.MakeSlice(pType, 0, 0)

				// Add default service (if exists)
				c.mu.RLock()
				if _, exists := c.services[elemType]; exists {
					c.mu.RUnlock()
					// Recursively resolve default instance
					inst, err := c.resolve(elemType, track)
					if err == nil {
						results = reflect.Append(results, inst)
					}
				} else {
					c.mu.RUnlock()
				}

				// Add all named services
				c.mu.RLock()
				for _, namedMap := range c.namedServices {
					if namedServiceDef, exists := namedMap[elemType]; exists {
						if namedServiceDef.isInstance {
							results = reflect.Append(results, namedServiceDef.instance)
						}
					}
				}
				c.mu.RUnlock()

				params[i] = results
			}
		} else if pType.Kind() == reflect.Map && pType.Key().Kind() == reflect.String {
			// Check if parameter is map[string]T type
			// First try to resolve map type directly (if registered)
			c.mu.RLock()
			_, mapExists := c.services[pType]
			c.mu.RUnlock()

			if mapExists {
				// map type is registered, resolve directly
				pInstance, err := c.resolve(pType, track)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
				}
				params[i] = pInstance
			} else {
				// map type not registered: automatically collect all named registered instances
				valueType := pType.Elem()

				// Create result map
				results := reflect.MakeMap(pType)

				// Collect all named services
				c.mu.RLock()
				for name, namedMap := range c.namedServices {
					if namedServiceDef, exists := namedMap[valueType]; exists {
						if namedServiceDef.isInstance {
							keyVal := reflect.ValueOf(name)
							results.SetMapIndex(keyVal, namedServiceDef.instance)
						}
					}
				}
				c.mu.RUnlock()

				params[i] = results
			}
		} else {
			// Non-slice/map type: normal resolution
			pInstance, err := c.resolve(pType, track)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
			}
			params[i] = pInstance
		}
	}

	// Call constructor to create instance
	results := serviceDef.ctor.Call(params)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("%w, constructor call returned abnormal value", ErrCreateInstanceFailed)
	}
	instance := results[0]

	// Singleton: atomic operation to cache instance, ensure created only once
	if serviceDef.scope == Singleton {
		serviceDef.once.Do(func() {
			serviceDef.instance = instance
		})
	}

	return instance, nil
}

// NewScope New: Container creates scope method (root container exclusive, creates Scoped scope)
func (c *Container) NewScope() *Scope {
	return &Scope{
		root:       c,
		scopedInst: make(map[reflect.Type]reflect.Value),
	}
}

// Resolve New: Scope's Resolve method (consistent format with Container's Resolve, supports Scoped)
func (s *Scope) Resolve(out any) error {
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr || outVal.IsNil() {
		return ErrInvalidOutPtr
	}
	svcType := outVal.Elem().Type()
	instance, err := s.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return err
	}
	outVal.Elem().Set(instance)
	return nil
}

// New: Scope's internal resolution method (handles all lifetimes, core Scoped caching logic)
func (s *Scope) resolve(svcType reflect.Type, track map[reflect.Type]bool) (reflect.Value, error) {
	// Get registration metadata from root container (shared by all scopes)
	s.root.mu.RLock()
	serviceDef, exists := s.root.services[svcType]
	s.root.mu.RUnlock()
	if !exists {
		return reflect.Value{}, fmt.Errorf("%w, type: %s", ErrServiceNotRegistered, svcType)
	}

	// Circular dependency detection
	if track[svcType] {
		return reflect.Value{}, fmt.Errorf("%w, circular dependency chain contains: %s", ErrResolveCircularDependency, svcType)
	}
	track[svcType] = true
	defer delete(track, svcType)

	// Instance registration handling
	if serviceDef.isInstance {
		// Singleton instance: directly return root container's instance
		if serviceDef.scope == Singleton {
			return serviceDef.instance, nil
		}
		// Scoped instance: each scope has independent cache
		if serviceDef.scope == Scoped {
			s.mu.RLock()
			inst, exists := s.scopedInst[svcType]
			s.mu.RUnlock()
			if exists && inst.IsValid() {
				return inst, nil
			}
			// First access: cache instance to scope
			s.mu.Lock()
			s.scopedInst[svcType] = serviceDef.instance
			s.mu.Unlock()
			return serviceDef.instance, nil
		}
	}

	// 1. Singleton: fix circular dependency → prioritize getting cache from root container, if not initialized use scope's own resolve (reuse track)
	if serviceDef.scope == Singleton {
		// Read lock to get root container's singleton instance, return directly if cached (core: skip root container resolve, avoid duplicate track writes)
		s.root.mu.RLock()
		if serviceDef.instance.IsValid() {
			inst := serviceDef.instance
			s.root.mu.RUnlock()
			return inst, nil
		}
		s.root.mu.RUnlock()
		// Singleton not initialized: use scope's own resolve to complete initialization (reuse current track, no circular dependency false positive)
		goto createInstance
	}

	// 2. Scoped: unique within scope, check this scope's cache first
	if serviceDef.scope == Scoped {
		s.mu.RLock()
		inst, exists := s.scopedInst[svcType]
		s.mu.RUnlock()
		if exists && inst.IsValid() {
			return inst, nil
		}
	}

	// New label: unified instance creation (Scoped/Transient/uninitialized Singleton shared)
createInstance:
	// Cache miss: resolve parameters + create instance (Scoped/Transient/uninitialized Singleton common)
	serviceDef.paramOnce.Do(func() {
		numIn := serviceDef.ctorType.NumIn()
		params := make([]reflect.Type, numIn)
		for i := 0; i < numIn; i++ {
			params[i] = serviceDef.ctorType.In(i)
		}
		serviceDef.paramTypes = params
	})
	paramTypes := serviceDef.paramTypes

	params := make([]reflect.Value, len(paramTypes))
	for i, pType := range paramTypes {
		// Check if parameter is a slice type
		if pType.Kind() == reflect.Slice {
			// First try to resolve slice type directly (if registered)
			s.root.mu.RLock()
			_, sliceExists := s.root.services[pType]
			s.root.mu.RUnlock()

			if sliceExists {
				// Slice type is registered, resolve directly
				pInstance, err := s.resolve(pType, track)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
				}
				params[i] = pInstance
			} else {
				// Slice type not registered: automatically collect all instances of that element type
				elemType := pType.Elem()

				// Create result slice
				results := reflect.MakeSlice(pType, 0, 0)

				// Add default service (if exists)
				s.root.mu.RLock()
				if _, exists := s.root.services[elemType]; exists {
					s.root.mu.RUnlock()
					// Recursively resolve default instance
					inst, err := s.resolve(elemType, track)
					if err == nil {
						results = reflect.Append(results, inst)
					}
				} else {
					s.root.mu.RUnlock()
				}

				// Add all named services
				s.root.mu.RLock()
				for _, namedMap := range s.root.namedServices {
					if namedServiceDef, exists := namedMap[elemType]; exists {
						if namedServiceDef.isInstance {
							results = reflect.Append(results, namedServiceDef.instance)
						}
					}
				}
				s.root.mu.RUnlock()

				params[i] = results
			}
		} else if pType.Kind() == reflect.Map && pType.Key().Kind() == reflect.String {
			// Check if parameter is map[string]T type
			// First try to resolve map type directly (if registered)
			s.root.mu.RLock()
			_, mapExists := s.root.services[pType]
			s.root.mu.RUnlock()

			if mapExists {
				// map type is registered, resolve directly
				pInstance, err := s.resolve(pType, track)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
				}
				params[i] = pInstance
			} else {
				// map type not registered: automatically collect all named registered instances
				valueType := pType.Elem()

				// Create result map
				results := reflect.MakeMap(pType)

				// Collect all named services
				s.root.mu.RLock()
				for name, namedMap := range s.root.namedServices {
					if namedServiceDef, exists := namedMap[valueType]; exists {
						if namedServiceDef.isInstance {
							keyVal := reflect.ValueOf(name)
							results.SetMapIndex(keyVal, namedServiceDef.instance)
						}
					}
				}
				s.root.mu.RUnlock()

				params[i] = results
			}
		} else {
			// Non-slice/map type: normal resolution
			pInstance, err := s.resolve(pType, track)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("failed to resolve dependency %s: %w", pType, err)
			}
			params[i] = pInstance
		}
	}

	results := serviceDef.ctor.Call(params)
	if len(results) != 1 {
		return reflect.Value{}, fmt.Errorf("%w, constructor call returned abnormal value", ErrCreateInstanceFailed)
	}
	instance := results[0]

	// 3. Scoped: write instance to this scope's cache
	if serviceDef.scope == Scoped {
		s.mu.Lock()
		s.scopedInst[svcType] = instance
		s.mu.Unlock()
	}

	// New: uninitialized Singleton, write to root container cache after creation (ensure global uniqueness)
	if serviceDef.scope == Singleton {
		serviceDef.once.Do(func() {
			s.root.mu.Lock()
			serviceDef.instance = instance
			s.root.mu.Unlock()
		})
	}

	// 4. Transient: return directly, no caching
	return instance, nil
}

// getTyped Internal generic resolution: converts reflection-obtained instance to target type T
func getTyped[T any](_ *Container, svcType reflect.Type, instance reflect.Value) (T, error) {
	var zero T
	// Handle interface types, assignable and convertible types
	it := instance.Type()
	// If target type is interface, check implementation relationship
	if svcType.Kind() == reflect.Interface {
		// Case 1: Instance type directly implements interface (including pointer types)
		if it.Implements(svcType) {
			return instance.Interface().(T), nil
		}
		// Case 2: Value type implements interface, but container returns value → try to get address
		if it.Kind() != reflect.Ptr && reflect.PointerTo(it).Implements(svcType) {
			var iface any
			if instance.CanAddr() {
				iface = instance.Addr().Interface()
			} else {
				// Create a new pointer and set value for conversion
				ptr := reflect.New(it)
				ptr.Elem().Set(instance)
				iface = ptr.Interface()
			}
			return iface.(T), nil
		}
		return zero, fmt.Errorf("[%w] instance %s cannot be converted to target interface type %s", ErrTypeConvertFailed, it, svcType)
	}

	// Target is not interface: check if directly assignable or convertible
	if it.AssignableTo(svcType) {
		return instance.Interface().(T), nil
	}
	if it.ConvertibleTo(svcType) {
		conv := instance.Convert(svcType)
		return conv.Interface().(T), nil
	}

	return zero, fmt.Errorf("[%w] instance %s cannot be converted to target type %s", ErrTypeConvertFailed, it, svcType)
}

// MustRegister ---------------------- Convenient Must series methods (panic on error, preferred for 90% scenarios) ----------------------
// MustRegister Convenient basic registration: panics directly on error
func (c *Container) MustRegister(ctor any, scope LifetimeScope) {
	if err := c.Register(ctor, scope); err != nil {
		panic(fmt.Sprintf("[DI Registration Failed] %v", err))
	}
}

// MustRegisterAs Convenient interface registration: panics directly on error
func (c *Container) MustRegisterAs(ctor any, interfaceType any, scope LifetimeScope) {
	if err := c.RegisterAs(ctor, interfaceType, scope); err != nil {
		panic(fmt.Sprintf("[DI Interface Registration Failed] %v", err))
	}
}

// MustRegisterInstance Convenient instance registration: panics directly on error
func (c *Container) MustRegisterInstance(instance any, scope LifetimeScope) {
	if err := c.RegisterInstance(instance, scope); err != nil {
		panic(fmt.Sprintf("[DI Instance Registration Failed] %v", err))
	}
}

// MustRegisterInstanceAs Convenient instance interface registration: panics directly on error
func (c *Container) MustRegisterInstanceAs(instance any, interfaceType any, scope LifetimeScope) {
	if err := c.RegisterInstanceAs(instance, interfaceType, scope); err != nil {
		panic(fmt.Sprintf("[DI Instance Interface Registration Failed] %v", err))
	}
}

// MustRegisterInstanceNamed Convenient named instance registration: panics directly on error
func (c *Container) MustRegisterInstanceNamed(name string, instance any, scope LifetimeScope) {
	if err := c.RegisterInstanceNamed(name, instance, scope); err != nil {
		panic(fmt.Sprintf("[DI Named Instance Registration Failed] %v", err))
	}
}

// MustRegisterInstanceAsNamed Convenient named instance interface registration: panics directly on error
func (c *Container) MustRegisterInstanceAsNamed(name string, instance any, interfaceType any, scope LifetimeScope) {
	if err := c.RegisterInstanceAsNamed(name, instance, interfaceType, scope); err != nil {
		panic(fmt.Sprintf("[DI Named Instance Interface Registration Failed] %v", err))
	}
}

// MustResolve Convenient original resolution: panics directly on error
func (c *Container) MustResolve(out any) {
	if err := c.Resolve(out); err != nil {
		panic(fmt.Sprintf("[DI Resolution Failed] %v", err))
	}
}

// MustResolveNamed Convenient named resolution: panics directly on error
func (c *Container) MustResolveNamed(name string, out any) {
	if err := c.ResolveNamed(name, out); err != nil {
		panic(fmt.Sprintf("[DI Named Resolution Failed] %v", err))
	}
}

// MustResolveAll Convenient resolve all: panics directly on error
func (c *Container) MustResolveAll(out any) {
	if err := c.ResolveAll(out); err != nil {
		panic(fmt.Sprintf("[DI Resolve All Failed] %v", err))
	}
}

// MustResolve New: Scope's MustResolve method (consistent format with Container)
func (s *Scope) MustResolve(out any) {
	if err := s.Resolve(out); err != nil {
		panic(fmt.Sprintf("[DI Scope Resolution Failed] %v", err))
	}
}

// MustRegister ---------------------- Global container top-level generic functions (directly call di.Get[T](), di.MustGet[T](), extremely concise) ----------------------
func MustRegister(ctor any, scope LifetimeScope) { Global.MustRegister(ctor, scope) }
func MustRegisterAs(ctor any, iface any, scope LifetimeScope) {
	Global.MustRegisterAs(ctor, iface, scope)
}
func MustRegisterInstance(instance any, scope LifetimeScope) {
	Global.MustRegisterInstance(instance, scope)
}
func MustRegisterInstanceAs(instance any, iface any, scope LifetimeScope) {
	Global.MustRegisterInstanceAs(instance, iface, scope)
}
func MustResolve(out any) { Global.MustResolve(out) }

// Get Generic resolution: directly returns instance with error handling (follows Go conventions)
func Get[T any]() (T, error) {
	var zero T
	svcType := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := Global.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return zero, fmt.Errorf("[DI Get Failed] %w", err)
	}
	return getTyped[T](Global, svcType, instance)
}

// MustGet Generic convenient resolution: directly returns instance, panics on error (recommended)
func MustGet[T any]() T {
	inst, err := Get[T]()
	if err != nil {
		panic(err)
	}
	return inst
}

// GlobalNewScope New: convenient method for creating scope globally
func GlobalNewScope() *Scope {
	return Global.NewScope()
}

// ScopeGet New: scope version generic Get - pass Scope pointer, implements Scoped lifetime generic resolution
func ScopeGet[T any](s *Scope) (T, error) {
	var zero T
	svcType := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := s.resolve(svcType, make(map[reflect.Type]bool))
	if err != nil {
		return zero, fmt.Errorf("[DI Scope Get Failed] %w", err)
	}
	return getTyped[T](s.root, svcType, instance)
}

// ScopeMustGet New: scope version generic MustGet - pass Scope pointer, panics on error (recommended)
func ScopeMustGet[T any](s *Scope) T {
	inst, err := ScopeGet[T](s)
	if err != nil {
		panic(err)
	}
	return inst
}

// Reset Resets container: clears all services and caches (for testing)
func (c *Container) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services = make(map[reflect.Type]*ServiceDef)
}

// Reset Replace with 👇 fixed code
func (s *Scope) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock() // Correct: use scope's own lock
	s.scopedInst = make(map[reflect.Type]reflect.Value)
}

// GlobalReset Resets global container (for testing)
func GlobalReset() { Global.Reset() }

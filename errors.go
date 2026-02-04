package gofac

import "errors"

// Framework core error definitions (added Scoped-related errors, original errors retained)
var (
	ErrNotFunc                   = errors.New("registration must be a constructor function (function type)")
	ErrNoReturn                  = errors.New("constructor must have exactly one return value")
	ErrRegisterDuplicate         = errors.New("service type already registered, duplicate registration prohibited")
	ErrServiceNotRegistered      = errors.New("service not registered, cannot resolve")
	ErrCreateInstanceFailed      = errors.New("failed to create service instance")
	ErrNotConcreteType           = errors.New("constructor return value must be concrete type (not interface)")
	ErrResolveCircularDependency = errors.New("circular dependency detected during resolution")
	ErrInvalidInterfaceType      = errors.New("interfaceType must be a nil pointer to interface, e.g. (*IInterface)(nil)")
	ErrInvalidOutPtr             = errors.New("out must be a non-nil pointer type")
	ErrTypeConvertFailed         = errors.New("instance cannot be converted to target type")
	ErrScopedOnRootContainer     = errors.New("scoped lifetime services cannot be retrieved directly from root container, please use Scope") // New Scoped error
	ErrTransientInstance         = errors.New("instance registration does not support Transient lifetime, please use Singleton or Scoped")
	ErrNilInstance               = errors.New("registered instance cannot be nil")
)

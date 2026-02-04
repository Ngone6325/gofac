package gofac

import (
	"errors"
	"testing"
)

// TestErrorConstants tests that all error constants are defined
func TestErrorConstants(t *testing.T) {
	errorTests := []struct {
		name  string
		err   error
		isNil bool
	}{
		{"ErrNotFunc", ErrNotFunc, false},
		{"ErrNoReturn", ErrNoReturn, false},
		{"ErrRegisterDuplicate", ErrRegisterDuplicate, false},
		{"ErrServiceNotRegistered", ErrServiceNotRegistered, false},
		{"ErrCreateInstanceFailed", ErrCreateInstanceFailed, false},
		{"ErrNotConcreteType", ErrNotConcreteType, false},
		{"ErrResolveCircularDependency", ErrResolveCircularDependency, false},
		{"ErrInvalidInterfaceType", ErrInvalidInterfaceType, false},
		{"ErrInvalidOutPtr", ErrInvalidOutPtr, false},
		{"ErrTypeConvertFailed", ErrTypeConvertFailed, false},
		{"ErrScopedOnRootContainer", ErrScopedOnRootContainer, false},
		{"ErrTransientInstance", ErrTransientInstance, false},
		{"ErrNilInstance", ErrNilInstance, false},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.err == nil) != tt.isNil {
				t.Errorf("%s: expected nil=%v, got nil=%v", tt.name, tt.isNil, tt.err == nil)
			}
		})
	}
}

// TestErrorMessages tests that error messages are not empty
func TestErrorMessages(t *testing.T) {
	errorTests := []error{
		ErrNotFunc,
		ErrNoReturn,
		ErrRegisterDuplicate,
		ErrServiceNotRegistered,
		ErrCreateInstanceFailed,
		ErrNotConcreteType,
		ErrResolveCircularDependency,
		ErrInvalidInterfaceType,
		ErrInvalidOutPtr,
		ErrTypeConvertFailed,
		ErrScopedOnRootContainer,
		ErrTransientInstance,
		ErrNilInstance,
	}

	for _, err := range errorTests {
		if err.Error() == "" {
			t.Errorf("Error message should not be empty for error: %v", err)
		}
	}
}

// TestErrorEquality tests that errors can be compared
func TestErrorEquality(t *testing.T) {
	if ErrNotFunc == ErrNoReturn {
		t.Error("Different errors should not be equal")
	}

	// Test with errors.Is
	testErr := ErrNotFunc
	if !errors.Is(testErr, ErrNotFunc) {
		t.Error("errors.Is should work with error constants")
	}

	if errors.Is(testErr, ErrNoReturn) {
		t.Error("errors.Is should return false for different errors")
	}
}

// TestErrorWrapping tests that errors can be wrapped
func TestErrorWrapping(t *testing.T) {
	wrappedErr := errors.Join(ErrNotFunc, errors.New("additional context"))

	if !errors.Is(wrappedErr, ErrNotFunc) {
		t.Error("Wrapped error should still be identifiable with errors.Is")
	}
}

// TestErrorTypes tests that all errors are of error type
func TestErrorTypes(t *testing.T) {
	var _ error = ErrNotFunc
	var _ error = ErrNoReturn
	var _ error = ErrRegisterDuplicate
	var _ error = ErrServiceNotRegistered
	var _ error = ErrCreateInstanceFailed
	var _ error = ErrNotConcreteType
	var _ error = ErrResolveCircularDependency
	var _ error = ErrInvalidInterfaceType
	var _ error = ErrInvalidOutPtr
	var _ error = ErrTypeConvertFailed
	var _ error = ErrScopedOnRootContainer
	var _ error = ErrTransientInstance
	var _ error = ErrNilInstance
}

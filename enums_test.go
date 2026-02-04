package gofac

import (
	"testing"
)

// TestLifetimeScopeConstants tests that the lifetime scope constants are defined correctly
func TestLifetimeScopeConstants(t *testing.T) {
	tests := []struct {
		name     string
		scope    LifetimeScope
		expected int
	}{
		{"Transient", Transient, 0},
		{"Singleton", Singleton, 1},
		{"Scoped", Scoped, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.scope) != tt.expected {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, int(tt.scope))
			}
		})
	}
}

// TestLifetimeScopeType tests that LifetimeScope is of type int
func TestLifetimeScopeType(t *testing.T) {
	var scope LifetimeScope = Transient
	_ = int(scope) // Should compile without error
}

// TestLifetimeScopeComparison tests that lifetime scopes can be compared
func TestLifetimeScopeComparison(t *testing.T) {
	if Transient == Singleton {
		t.Error("Transient should not equal Singleton")
	}

	if Singleton == Scoped {
		t.Error("Singleton should not equal Scoped")
	}

	if Transient == Scoped {
		t.Error("Transient should not equal Scoped")
	}

	// Test equality
	var scope1 LifetimeScope = Singleton
	var scope2 LifetimeScope = Singleton
	if scope1 != scope2 {
		t.Error("Same lifetime scopes should be equal")
	}
}

// TestLifetimeScopeOrdering tests the ordering of lifetime scope values
func TestLifetimeScopeOrdering(t *testing.T) {
	if Transient >= Singleton {
		t.Error("Transient should be less than Singleton")
	}

	if Singleton >= Scoped {
		t.Error("Singleton should be less than Scoped")
	}

	if Transient >= Scoped {
		t.Error("Transient should be less than Scoped")
	}
}

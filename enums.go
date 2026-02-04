package gofac

type LifetimeScope int

const (
	Transient LifetimeScope = iota // Transient: creates new instance on each retrieval
	Singleton                      // Singleton: globally unique, cached in root container
	Scoped                         // Scoped: unique within scope, isolated between different scopes
)

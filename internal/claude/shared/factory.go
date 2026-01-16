package shared

import (
	"sync"
)

// Factory is a generic interface for creating instances of type T with options of type O.
// This enables dependency injection and testability across packages.
type Factory[T any, O any] interface {
	// New creates a new instance of T with the given options.
	New(opts ...func(*O)) (T, error)
}

// FactoryFunc is a function type that implements Factory[T, O].
// This allows for easy inline factory creation.
type FactoryFunc[T any, O any] func(opts ...func(*O)) (T, error)

// New implements Factory[T, O].
func (f FactoryFunc[T, O]) New(opts ...func(*O)) (T, error) {
	return f(opts...)
}

// DefaultFactoryHolder holds a default factory instance with thread-safe access.
// Use this to implement package-level default factories.
type DefaultFactoryHolder[T any, O any] struct {
	mu      sync.RWMutex
	factory Factory[T, O]
	fallback Factory[T, O]
}

// NewDefaultFactoryHolder creates a new holder with the given fallback factory.
func NewDefaultFactoryHolder[T any, O any](fallback Factory[T, O]) *DefaultFactoryHolder[T, O] {
	return &DefaultFactoryHolder[T, O]{
		factory:  fallback,
		fallback: fallback,
	}
}

// Get returns the current factory.
func (h *DefaultFactoryHolder[T, O]) Get() Factory[T, O] {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.factory
}

// Set sets the factory, or resets to fallback if nil.
func (h *DefaultFactoryHolder[T, O]) Set(factory Factory[T, O]) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if factory == nil {
		h.factory = h.fallback
	} else {
		h.factory = factory
	}
}

// ConstructorFunc represents a function that constructs a T from options O.
// This is the actual creation logic that DefaultFactory wraps.
type ConstructorFunc[T any, O any] func(opts *O) (T, error)

// DefaultFactory creates a factory from a constructor and default options provider.
// This is a helper for creating simple factories without custom logic.
func DefaultFactory[T any, O any](constructor ConstructorFunc[T, O], defaults func() *O) Factory[T, O] {
	return FactoryFunc[T, O](func(opts ...func(*O)) (T, error) {
		options := defaults()
		for _, opt := range opts {
			opt(options)
		}
		return constructor(options)
	})
}

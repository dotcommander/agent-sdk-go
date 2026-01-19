package v2

import (
	"github.com/dotcommander/agent-sdk-go/claude"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// ClientFactory creates claude.Client instances.
// This interface enables dependency injection for session/prompt creation (DIP compliance).
// By depending on this abstraction rather than concrete NewClient, sessions can be:
//   - Tested with mock clients
//   - Configured with custom client implementations
//   - Extended without modifying session code
type ClientFactory interface {
	// NewClient creates a new Client with the given options.
	NewClient(opts ...claude.ClientOption) (claude.Client, error)
}

// defaultClientFactory is the standard factory that uses claude.NewClient.
type defaultClientFactory struct{}

// NewClient creates a new Client using the default implementation.
func (f *defaultClientFactory) NewClient(opts ...claude.ClientOption) (claude.Client, error) {
	return claude.NewClient(opts...)
}

// defaultFactoryHolder holds the package-level default factory using the generic holder.
var defaultFactoryHolder = shared.NewDefaultFactoryHolder[claude.Client, claude.ClientOptions](
	shared.DefaultFactory[claude.Client, claude.ClientOptions](
		func(opts *claude.ClientOptions) (claude.Client, error) {
			return claude.NewClient()
		},
		claude.DefaultClientOptions,
	),
)

// DefaultClientFactory returns the default client factory.
func DefaultClientFactory() ClientFactory {
	return &v2FactoryAdapter{defaultFactoryHolder.Get()}
}

// SetDefaultClientFactory sets the package-level default client factory.
// This is useful for testing or for providing a custom client implementation globally.
func SetDefaultClientFactory(factory ClientFactory) {
	if factory == nil {
		defaultFactoryHolder.Set(nil)
		return
	}
	// Wrap the ClientFactory to match the generic interface
	defaultFactoryHolder.Set(shared.FactoryFunc[claude.Client, claude.ClientOptions](func(opts ...func(*claude.ClientOptions)) (claude.Client, error) {
		clientOpts := make([]claude.ClientOption, len(opts))
		for i, opt := range opts {
			clientOpts[i] = opt
		}
		return factory.NewClient(clientOpts...)
	}))
}

// v2FactoryAdapter adapts the generic Factory to ClientFactory interface.
type v2FactoryAdapter struct {
	generic shared.Factory[claude.Client, claude.ClientOptions]
}

func (a *v2FactoryAdapter) NewClient(opts ...claude.ClientOption) (claude.Client, error) {
	optFuncs := make([]func(*claude.ClientOptions), len(opts))
	for i, opt := range opts {
		optFuncs[i] = opt
	}
	return a.generic.New(optFuncs...)
}

// ClientFactoryFunc is a function type that implements ClientFactory.
// This allows for easy inline factory creation.
//
// Example:
//
//	factory := v2.ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
//	    return myMockClient, nil
//	})
type ClientFactoryFunc func(opts ...claude.ClientOption) (claude.Client, error)

// NewClient implements ClientFactory.
func (f ClientFactoryFunc) NewClient(opts ...claude.ClientOption) (claude.Client, error) {
	return f(opts...)
}

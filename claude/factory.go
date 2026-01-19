package claude

import (
	"context"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// ClientFactory creates Client instances.
// This interface allows for dependency injection of client creation (DIP compliance).
// Sessions and other high-level APIs depend on this abstraction rather than concrete NewClient.
type ClientFactory interface {
	// NewClient creates a new Client with the given options.
	NewClient(opts ...ClientOption) (Client, error)
}

// DefaultClientFactory is the standard factory that creates ClientImpl instances.
type DefaultClientFactory struct{}

// NewClient creates a new Client using the default implementation.
func (f *DefaultClientFactory) NewClient(opts ...ClientOption) (Client, error) {
	return NewClient(opts...)
}

// defaultFactoryHolder holds the global default factory using the generic holder.
var defaultFactoryHolder = shared.NewDefaultFactoryHolder[Client, ClientOptions](
	shared.DefaultFactory[Client, ClientOptions](
		func(opts *ClientOptions) (Client, error) {
			return &ClientImpl{options: opts}, nil
		},
		DefaultClientOptions,
	),
)

// DefaultFactory returns the default client factory.
func DefaultFactory() ClientFactory {
	return &factoryAdapter{defaultFactoryHolder.Get()}
}

// SetDefaultFactory sets the global default client factory.
// This is useful for testing or for providing a custom client implementation globally.
func SetDefaultFactory(factory ClientFactory) {
	if factory == nil {
		defaultFactoryHolder.Set(nil)
		return
	}
	// Wrap the ClientFactory to match the generic interface
	defaultFactoryHolder.Set(shared.FactoryFunc[Client, ClientOptions](func(opts ...func(*ClientOptions)) (Client, error) {
		clientOpts := make([]ClientOption, len(opts))
		for i, opt := range opts {
			clientOpts[i] = opt
		}
		return factory.NewClient(clientOpts...)
	}))
}

// factoryAdapter adapts the generic Factory to ClientFactory interface.
type factoryAdapter struct {
	generic shared.Factory[Client, ClientOptions]
}

func (a *factoryAdapter) NewClient(opts ...ClientOption) (Client, error) {
	optFuncs := make([]func(*ClientOptions), len(opts))
	for i, opt := range opts {
		optFuncs[i] = opt
	}
	return a.generic.New(optFuncs...)
}

// ClientFactoryFunc is a function type that implements ClientFactory.
// This allows for easy inline factory creation.
type ClientFactoryFunc func(opts ...ClientOption) (Client, error)

// NewClient implements ClientFactory.
func (f ClientFactoryFunc) NewClient(opts ...ClientOption) (Client, error) {
	return f(opts...)
}

// ConnectedClientFactory wraps ClientFactory and automatically connects clients.
// This is useful when you always want clients to be connected upon creation.
type ConnectedClientFactory struct {
	Factory ClientFactory
}

// NewClient creates and connects a new Client.
func (f *ConnectedClientFactory) NewClient(opts ...ClientOption) (Client, error) {
	client, err := f.Factory.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	// Connect with a background context - caller should manage actual connection
	if err := client.Connect(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
}

package v2

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"agent-sdk-go/internal/claude"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient is a simple mock implementation of claude.Client for testing.
type MockClient struct {
	name  string
	model string
}

func (m *MockClient) Connect(ctx context.Context) error {
	return nil
}

func (m *MockClient) Disconnect() error {
	return nil
}

func (m *MockClient) Query(ctx context.Context, prompt string) (string, error) {
	return "mock response", nil
}

func (m *MockClient) QueryWithSession(ctx context.Context, sessionID string, prompt string) (string, error) {
	return "mock response", nil
}

func (m *MockClient) QueryStream(ctx context.Context, prompt string) (<-chan claude.Message, <-chan error) {
	msgChan := make(chan claude.Message)
	errChan := make(chan error)
	close(msgChan)
	close(errChan)
	return msgChan, errChan
}

func (m *MockClient) ReceiveMessages(ctx context.Context) (<-chan claude.Message, <-chan error) {
	msgChan := make(chan claude.Message)
	errChan := make(chan error)
	close(msgChan)
	close(errChan)
	return msgChan, errChan
}

func (m *MockClient) ReceiveResponse(ctx context.Context) (claude.Message, error) {
	return nil, nil
}

func (m *MockClient) Interrupt() error {
	return nil
}

func (m *MockClient) SetModel(model string) {
	m.model = model
}

func (m *MockClient) SetPermissionMode(mode string) {
	// Mock implementation
}

func (m *MockClient) RewindFiles(ctx context.Context, files []string) error {
	return nil
}

func (m *MockClient) GetOptions() *claude.ClientOptions {
	opts := claude.DefaultClientOptions()
	opts.Model = m.model
	return opts
}

// Ensure MockClient implements claude.Client interface
var _ claude.Client = (*MockClient)(nil)

// TestDefaultClientFactory tests the default factory implementation.
func TestDefaultClientFactory(t *testing.T) {
	t.Parallel()

	factory := &defaultClientFactory{}

	// Create a client with options
	client, err := factory.NewClient(claude.WithModel("claude-3-5-sonnet-20241022"))
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestClientFactoryFunc tests the function-based factory adapter.
func TestClientFactoryFunc(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{
		name:  "test",
		model: "claude-3-5-sonnet-20241022",
	}

	factory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return mockClient, nil
	})

	client, err := factory.NewClient()
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, mockClient, client)
}

// TestDefaultClientFactoryGetter tests getting the default factory.
func TestDefaultClientFactoryGetter(t *testing.T) {
	t.Parallel()

	factory := DefaultClientFactory()
	require.NotNil(t, factory)

	// Should be able to call NewClient on it
	client, err := factory.NewClient(claude.WithModel("claude-3-5-sonnet-20241022"))
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestSetDefaultClientFactory tests setting a custom default factory.
func TestSetDefaultClientFactory(t *testing.T) {
	t.Parallel()

	// Save original
	originalFactory := DefaultClientFactory()
	defer SetDefaultClientFactory(originalFactory)

	// Create a mock factory
	mockClient := &MockClient{
		name:  "custom",
		model: "test-model",
	}
	customFactory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return mockClient, nil
	})

	// Set it
	SetDefaultClientFactory(customFactory)

	// Get it back
	factory := DefaultClientFactory()
	require.NotNil(t, factory)

	// Use it
	client, err := factory.NewClient()
	require.NoError(t, err)
	assert.Equal(t, mockClient, client)
}

// TestSetDefaultClientFactoryNil tests that nil factory falls back to default.
func TestSetDefaultClientFactoryNil(t *testing.T) {
	t.Parallel()

	// Save original
	originalFactory := DefaultClientFactory()
	defer SetDefaultClientFactory(originalFactory)

	// Set to nil - should use default
	SetDefaultClientFactory(nil)

	factory := DefaultClientFactory()
	require.NotNil(t, factory)

	// Should work and return a valid client
	client, err := factory.NewClient(claude.WithModel("claude-3-5-sonnet-20241022"))
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestClientFactoryFuncImplementsInterface verifies ClientFactoryFunc implements ClientFactory.
func TestClientFactoryFuncImplementsInterface(t *testing.T) {
	t.Parallel()

	factory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return nil, nil
	})

	// Should be assignable to ClientFactory
	var _ ClientFactory = factory
}

// TestClientFactoryWithOptions tests factory with various options.
func TestClientFactoryWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []claude.ClientOption
		wantErr bool
	}{
		{
			name:    "no options",
			opts:    []claude.ClientOption{},
			wantErr: false,
		},
		{
			name:    "with model",
			opts:    []claude.ClientOption{claude.WithModel("claude-3-5-sonnet-20241022")},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &defaultClientFactory{}
			client, err := factory.NewClient(tt.opts...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

// TestClientFactoryThreadSafety tests concurrent factory usage.
func TestClientFactoryThreadSafety(t *testing.T) {
	t.Parallel()

	mockClient := &MockClient{
		name:  "concurrent",
		model: "test-model",
	}
	factory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return mockClient, nil
	})

	var wg sync.WaitGroup
	results := make([]error, 10)

	for i := range 10 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, results[index] = factory.NewClient()
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i, err := range results {
		assert.NoError(t, err, "concurrent call %d should succeed", i)
	}
}

// TestSetDefaultClientFactoryThreadSafety tests concurrent factory setting.
func TestSetDefaultClientFactoryThreadSafety(t *testing.T) {
	t.Parallel()

	// Save original
	originalFactory := DefaultClientFactory()
	defer SetDefaultClientFactory(originalFactory)

	var wg sync.WaitGroup

	// Concurrently set and get factory
	for i := range 5 {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			mockClient := &MockClient{
				name:  fmt.Sprintf("concurrent_%d", index),
				model: "test-model",
			}
			factory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
				return mockClient, nil
			})

			SetDefaultClientFactory(factory)
		}(i)
	}

	wg.Wait()

	// Factory should still be set and usable
	factory := DefaultClientFactory()
	require.NotNil(t, factory)
}

// TestClientFactoryDIP tests dependency injection compliance (DIP).
func TestClientFactoryDIP(t *testing.T) {
	t.Parallel()

	// This test verifies that code depends on ClientFactory interface, not concrete impl

	// Create a mock client
	mockClient := &MockClient{
		name:  "dip",
		model: "test-model",
	}

	// Create a factory
	var factory ClientFactory = ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return mockClient, nil
	})

	// Code that accepts factory should work with any implementation
	client, err := factory.NewClient(claude.WithModel("claude-3-5-sonnet-20241022"))
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestClientFactoryInSessionCreation tests factory usage in session creation context.
func TestClientFactoryInSessionCreation(t *testing.T) {
	t.Parallel()

	// Save original
	originalFactory := DefaultClientFactory()
	defer SetDefaultClientFactory(originalFactory)

	// Create a mock client for session
	mockClient := &MockClient{
		name:  "session",
		model: "test-model",
	}

	// Create a factory that returns our mock
	customFactory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return mockClient, nil
	})

	// Set as default
	SetDefaultClientFactory(customFactory)

	// Now when CreateSession is called with factory option, it should use our mock
	factory := DefaultClientFactory()
	require.NotNil(t, factory)

	client, err := factory.NewClient()
	require.NoError(t, err)
	assert.Equal(t, mockClient, client)
}

// BenchmarkClientFactory benchmarks factory creation.
func BenchmarkClientFactory(b *testing.B) {
	factory := ClientFactoryFunc(func(opts ...claude.ClientOption) (claude.Client, error) {
		return &MockClient{
			name:  "bench",
			model: "test-model",
		}, nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := factory.NewClient()
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkDefaultClientFactory benchmarks the default factory.
func BenchmarkDefaultClientFactory(b *testing.B) {
	factory := DefaultClientFactory()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = factory.NewClient(claude.WithModel("claude-3-5-sonnet-20241022"))
			// Ignore error - just benchmarking the interface
		}
	})
}

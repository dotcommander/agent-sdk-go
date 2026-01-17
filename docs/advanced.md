# Advanced Usage

This guide covers low-level SDK features for advanced use cases.

## Direct Transport Access

For maximum control, use the transport layer directly:

```go
import "github.com/dotcommander/agent-sdk-go/claude/subprocess"

config := &subprocess.TransportConfig{
    Model:   "claude-sonnet-4-20250514",
    Timeout: 60 * time.Second,
}

// One-shot with prompt
transport, err := subprocess.NewTransportWithPrompt(config, "Hello!")
if err != nil {
    log.Fatal(err)
}

err = transport.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer transport.Close()

msgChan, errChan := transport.ReceiveMessages(ctx)
for msg := range msgChan {
    fmt.Printf("%T: %+v\n", msg, msg)
}
```

## Custom Message Parsing

Process raw JSON messages:

```go
import "github.com/dotcommander/agent-sdk-go/claude/parser"

// Create a parser
p := parser.NewJSONParser()

// Parse raw JSON
jsonData := []byte(`{"type": "assistant", "content": [...]}`)
msg, err := p.Parse(jsonData)
if err != nil {
    log.Fatal(err)
}

// Type switch on result
switch m := msg.(type) {
case *parser.AssistantMessage:
    fmt.Println("Assistant:", m.Content)
case *parser.ToolUseMessage:
    fmt.Println("Tool:", m.Name)
}
```

## Building Custom Clients

Create a custom client wrapper:

```go
type MyClient struct {
    *claude.ClientImpl
    logger *log.Logger
    metrics *Metrics
}

func NewMyClient(opts ...claude.ClientOption) (*MyClient, error) {
    baseClient, err := claude.NewClient(opts...)
    if err != nil {
        return nil, err
    }

    return &MyClient{
        ClientImpl: baseClient.(*claude.ClientImpl),
        logger:     log.New(os.Stdout, "[claude] ", log.LstdFlags),
        metrics:    NewMetrics(),
    }, nil
}

func (c *MyClient) Query(ctx context.Context, prompt string) (string, error) {
    start := time.Now()
    c.logger.Printf("Query: %s", prompt[:min(50, len(prompt))])

    response, err := c.ClientImpl.Query(ctx, prompt)

    c.metrics.RecordLatency(time.Since(start))
    if err != nil {
        c.metrics.RecordError()
        c.logger.Printf("Error: %v", err)
    } else {
        c.metrics.RecordSuccess()
    }

    return response, err
}
```

## Message Interceptors

Intercept and modify messages:

```go
type InterceptingClient struct {
    claude.Client
    beforeSend func(string) string
    afterRecv  func(claude.Message) claude.Message
}

func (c *InterceptingClient) Query(ctx context.Context, prompt string) (string, error) {
    // Modify prompt before sending
    if c.beforeSend != nil {
        prompt = c.beforeSend(prompt)
    }

    return c.Client.Query(ctx, prompt)
}

func (c *InterceptingClient) QueryStream(ctx context.Context, prompt string) (<-chan claude.Message, <-chan error) {
    if c.beforeSend != nil {
        prompt = c.beforeSend(prompt)
    }

    msgChan, errChan := c.Client.QueryStream(ctx, prompt)

    // Wrap message channel for interception
    if c.afterRecv != nil {
        outChan := make(chan claude.Message)
        go func() {
            defer close(outChan)
            for msg := range msgChan {
                outChan <- c.afterRecv(msg)
            }
        }()
        return outChan, errChan
    }

    return msgChan, errChan
}
```

## Pool of Clients

Manage a pool for high-throughput:

```go
type ClientPool struct {
    clients chan claude.Client
    factory func() (claude.Client, error)
}

func NewClientPool(size int, opts ...claude.ClientOption) (*ClientPool, error) {
    pool := &ClientPool{
        clients: make(chan claude.Client, size),
        factory: func() (claude.Client, error) {
            return claude.NewClient(opts...)
        },
    }

    // Pre-create clients
    for i := 0; i < size; i++ {
        client, err := pool.factory()
        if err != nil {
            return nil, err
        }
        pool.clients <- client
    }

    return pool, nil
}

func (p *ClientPool) Acquire() claude.Client {
    return <-p.clients
}

func (p *ClientPool) Release(c claude.Client) {
    p.clients <- c
}

func (p *ClientPool) Query(ctx context.Context, prompt string) (string, error) {
    client := p.Acquire()
    defer p.Release(client)
    return client.Query(ctx, prompt)
}
```

## Custom Error Handling

Implement retry with backoff:

```go
type RetryConfig struct {
    MaxAttempts int
    InitialWait time.Duration
    MaxWait     time.Duration
    Multiplier  float64
}

func QueryWithRetry(client claude.Client, ctx context.Context, prompt string, cfg RetryConfig) (string, error) {
    var lastErr error
    wait := cfg.InitialWait

    for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
        response, err := client.Query(ctx, prompt)
        if err == nil {
            return response, nil
        }

        lastErr = err

        // Don't retry on context errors
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            return "", err
        }

        // Wait before retry
        select {
        case <-time.After(wait):
            wait = time.Duration(float64(wait) * cfg.Multiplier)
            if wait > cfg.MaxWait {
                wait = cfg.MaxWait
            }
        case <-ctx.Done():
            return "", ctx.Err()
        }
    }

    return "", fmt.Errorf("failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
}
```

## Structured Output Parsing

Parse Claude's response into structured data:

```go
type CodeReview struct {
    Issues   []Issue `json:"issues"`
    Score    int     `json:"score"`
    Summary  string  `json:"summary"`
}

type Issue struct {
    Line    int    `json:"line"`
    Type    string `json:"type"`
    Message string `json:"message"`
}

func ReviewCode(client claude.Client, ctx context.Context, code string) (*CodeReview, error) {
    prompt := fmt.Sprintf(`Review this code and respond with JSON only:
{
    "issues": [{"line": N, "type": "bug|style|perf", "message": "..."}],
    "score": 1-10,
    "summary": "..."
}

Code:
%s`, code)

    response, err := client.Query(ctx, prompt)
    if err != nil {
        return nil, err
    }

    // Extract JSON from response
    response = extractJSON(response)

    var review CodeReview
    if err := json.Unmarshal([]byte(response), &review); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }

    return &review, nil
}

func extractJSON(s string) string {
    start := strings.Index(s, "{")
    end := strings.LastIndex(s, "}")
    if start >= 0 && end > start {
        return s[start : end+1]
    }
    return s
}
```

## Middleware Pattern

Apply middleware to all requests:

```go
type Middleware func(next QueryFunc) QueryFunc
type QueryFunc func(ctx context.Context, prompt string) (string, error)

func ChainMiddleware(client claude.Client, middlewares ...Middleware) QueryFunc {
    // Start with base query
    query := client.Query

    // Apply middlewares in reverse order
    for i := len(middlewares) - 1; i >= 0; i-- {
        query = middlewares[i](query)
    }

    return query
}

// Logging middleware
func LoggingMiddleware(logger *log.Logger) Middleware {
    return func(next QueryFunc) QueryFunc {
        return func(ctx context.Context, prompt string) (string, error) {
            logger.Printf("Query: %s", prompt[:min(50, len(prompt))])
            start := time.Now()
            resp, err := next(ctx, prompt)
            logger.Printf("Response in %v", time.Since(start))
            return resp, err
        }
    }
}

// Caching middleware
func CachingMiddleware(cache *Cache) Middleware {
    return func(next QueryFunc) QueryFunc {
        return func(ctx context.Context, prompt string) (string, error) {
            key := hash(prompt)
            if cached, ok := cache.Get(key); ok {
                return cached, nil
            }
            resp, err := next(ctx, prompt)
            if err == nil {
                cache.Set(key, resp)
            }
            return resp, err
        }
    }
}

// Usage
query := ChainMiddleware(client,
    LoggingMiddleware(logger),
    CachingMiddleware(cache),
)
response, err := query(ctx, "Hello")
```

## Testing with Mocks

Create testable code:

```go
// Interface for testing
type QueryService interface {
    Query(ctx context.Context, prompt string) (string, error)
}

// Production implementation
type ClaudeService struct {
    client claude.Client
}

func (s *ClaudeService) Query(ctx context.Context, prompt string) (string, error) {
    return s.client.Query(ctx, prompt)
}

// Mock for testing
type MockService struct {
    responses map[string]string
}

func (m *MockService) Query(ctx context.Context, prompt string) (string, error) {
    if resp, ok := m.responses[prompt]; ok {
        return resp, nil
    }
    return "", errors.New("unexpected prompt")
}

// Test
func TestMyFeature(t *testing.T) {
    mock := &MockService{
        responses: map[string]string{
            "What is 2+2?": "4",
        },
    }

    result := MyFeature(mock)
    assert.Equal(t, expected, result)
}
```

## Performance Monitoring

Track SDK performance:

```go
type Metrics struct {
    mu           sync.Mutex
    totalQueries int64
    totalErrors  int64
    totalLatency time.Duration
}

func (m *Metrics) Record(latency time.Duration, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.totalQueries++
    m.totalLatency += latency
    if err != nil {
        m.totalErrors++
    }
}

func (m *Metrics) Stats() (queries, errors int64, avgLatency time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()

    queries = m.totalQueries
    errors = m.totalErrors
    if queries > 0 {
        avgLatency = m.totalLatency / time.Duration(queries)
    }
    return
}
```

## Next Steps

- [Troubleshooting](troubleshooting.md) - Debug issues
- [Configuration](configuration.md) - All options

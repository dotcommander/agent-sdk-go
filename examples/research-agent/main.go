// Package main demonstrates a multi-agent research system with the Claude Agent SDK.
//
// This example shows:
// - Multi-agent coordination patterns
// - Subagent spawning and management
// - Research workflow orchestration
// - Result aggregation from multiple sources
// - PDF report generation concepts
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/research-agent
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	fmt.Println("=== Multi-Agent Research System ===")
	fmt.Println()

	// Check CLI availability
	if !cli.IsCLIAvailable() {
		fmt.Println("Claude CLI not available.")
		demonstratePatterns()
		return
	}

	// Get research topic
	topic := "artificial intelligence in healthcare"
	if len(os.Args) > 1 {
		topic = strings.Join(os.Args[1:], " ")
	}

	fmt.Printf("Research topic: %s\n", topic)
	fmt.Println()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Run research
	runResearch(ctx, topic)
}

// demonstratePatterns shows multi-agent patterns without actual execution.
func demonstratePatterns() {
	fmt.Println("--- Multi-Agent Research Patterns ---")
	fmt.Println()

	fmt.Println("1. Research Orchestrator:")
	fmt.Print(`
  // Main orchestrator session
  orchestrator, _ := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithSystemPrompt(` + "`" + `You are a research orchestrator.
      Coordinate specialized agents to research topics thoroughly.` + "`" + `),
  )
`)
	fmt.Println()

	fmt.Println("2. Specialized Subagents:")
	fmt.Print(`
  // Web research agent
  webAgent, _ := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithAllowedTools([]string{"WebSearch", "WebFetch"}),
      v2.WithSystemPrompt("You research topics using web searches."),
  )

  // Code analysis agent
  codeAgent, _ := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithAllowedTools([]string{"Read", "Glob", "Grep"}),
      v2.WithSystemPrompt("You analyze code and technical documentation."),
  )

  // Writing agent
  writeAgent, _ := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithSystemPrompt("You synthesize research into clear reports."),
  )
`)
	fmt.Println()

	fmt.Println("3. Parallel Research Pattern:")
	fmt.Print(`
  var wg sync.WaitGroup
  results := make(chan ResearchResult, 3)

  // Launch parallel research tasks
  wg.Add(3)
  go func() {
      defer wg.Done()
      results <- webAgent.Research(ctx, topic)
  }()
  go func() {
      defer wg.Done()
      results <- codeAgent.Research(ctx, topic)
  }()
  go func() {
      defer wg.Done()
      results <- academicAgent.Research(ctx, topic)
  }()

  // Wait and collect
  go func() {
      wg.Wait()
      close(results)
  }()

  for result := range results {
      allResults = append(allResults, result)
  }
`)
	fmt.Println()

	fmt.Println("4. Result Synthesis:")
	synthesisCode := `
  // Combine results into final report
  synthesisPrompt := fmt.Sprintf(` + "`" + `
  Synthesize these research findings into a comprehensive report:

  Web Research:
  ` + "`" + `+"%s"+` + "`" + `

  Code Analysis:
  ` + "`" + `+"%s"+` + "`" + `

  Academic Sources:
  ` + "`" + `+"%s"+` + "`" + `

  Create a well-structured report with:
  - Executive Summary
  - Key Findings
  - Detailed Analysis
  - Conclusions
  - References
  ` + "`" + `, webResults, codeResults, academicResults)

  writeAgent.Send(ctx, synthesisPrompt)
`
	fmt.Print(synthesisCode)
	fmt.Println()
}

// runResearch executes the research workflow.
func runResearch(ctx context.Context, topic string) {
	fmt.Println("Starting research workflow...")
	fmt.Println()

	// Create orchestrator session
	orchestrator, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(5*time.Minute),
	)
	if err != nil {
		fmt.Printf("Failed to create orchestrator: %v\n", err)
		return
	}
	defer func() { _ = orchestrator.Close() }()

	// Research phases
	phases := []struct {
		name   string
		prompt string
	}{
		{
			name: "1. Topic Analysis",
			prompt: fmt.Sprintf(`Analyze this research topic and break it into key subtopics:

Topic: %s

List 3-5 key areas to research, each with:
- Subtopic name
- Key questions to answer
- Suggested research approach`, topic),
		},
		{
			name: "2. Key Findings",
			prompt: fmt.Sprintf(`Based on your knowledge, provide key findings about:

Topic: %s

Include:
- Current state of the field
- Major developments and trends
- Key players and stakeholders
- Challenges and opportunities`, topic),
		},
		{
			name: "3. Synthesis",
			prompt: `Based on the analysis above, create a brief executive summary that:
- Highlights the most important insights
- Identifies key trends
- Suggests areas for further research

Keep it concise (2-3 paragraphs).`,
		},
	}

	// Execute each phase
	for _, phase := range phases {
		fmt.Printf("--- %s ---\n", phase.name)

		if err := orchestrator.Send(ctx, phase.prompt); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Stream response
		for msg := range orchestrator.Receive(ctx) {
			switch msg.Type() {
			case v2.V2EventTypeAssistant:
				fmt.Print(v2.ExtractAssistantText(msg))
			case v2.V2EventTypeStreamDelta:
				fmt.Print(v2.ExtractDeltaText(msg))
			case v2.V2EventTypeResult:
				// Phase complete
			case v2.V2EventTypeError:
				fmt.Printf("\n[Error: %s]\n", v2.ExtractErrorMessage(msg))
				return
			}
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Println("=== Research Complete ===")
}

// ResearchResult represents a result from a research agent.
type ResearchResult struct {
	AgentType string
	Topic     string
	Findings  []string
	Sources   []string
	Error     error
}

// ResearchAgent represents a specialized research agent.
type ResearchAgent struct {
	Name         string
	Session      v2.V2Session
	Specialty    string
	AllowedTools []string
}

// NewResearchAgent creates a new research agent.
func NewResearchAgent(ctx context.Context, name, specialty string, tools []string) (*ResearchAgent, error) {
	systemPrompt := fmt.Sprintf(`You are a specialized research agent focused on %s.
Your role is to gather and analyze information in your area of expertise.
Be thorough but concise in your findings.`, specialty)

	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(2*time.Minute),
		v2.WithSystemPrompt(systemPrompt),
	)
	if err != nil {
		return nil, err
	}

	return &ResearchAgent{
		Name:         name,
		Session:      session,
		Specialty:    specialty,
		AllowedTools: tools,
	}, nil
}

// Research performs research on a topic.
func (ra *ResearchAgent) Research(ctx context.Context, topic string) ResearchResult {
	result := ResearchResult{
		AgentType: ra.Name,
		Topic:     topic,
	}

	prompt := fmt.Sprintf(`Research the following topic from your specialty perspective (%s):

Topic: %s

Provide:
1. Key findings (3-5 bullet points)
2. Relevant sources or references
3. Any limitations or gaps in the research`, ra.Specialty, topic)

	if err := ra.Session.Send(ctx, prompt); err != nil {
		result.Error = err
		return result
	}

	var response strings.Builder
	for msg := range ra.Session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			response.WriteString(v2.ExtractAssistantText(msg))
		case v2.V2EventTypeStreamDelta:
			response.WriteString(v2.ExtractDeltaText(msg))
		case v2.V2EventTypeError:
			result.Error = fmt.Errorf(v2.ExtractErrorMessage(msg))
			return result
		}
	}

	result.Findings = []string{response.String()}
	return result
}

// Close closes the research agent.
func (ra *ResearchAgent) Close() error {
	return ra.Session.Close()
}

// ParallelResearch runs multiple research agents in parallel.
func ParallelResearch(ctx context.Context, agents []*ResearchAgent, topic string) []ResearchResult {
	var wg sync.WaitGroup
	results := make([]ResearchResult, len(agents))

	for i, agent := range agents {
		wg.Add(1)
		go func(idx int, a *ResearchAgent) {
			defer wg.Done()
			results[idx] = a.Research(ctx, topic)
		}(i, agent)
	}

	wg.Wait()
	return results
}

// Package main demonstrates structured output with the Claude Agent SDK.
//
// Structured output allows you to:
// - Get responses in a specific JSON schema
// - Parse responses into Go structs
// - Ensure type-safe data extraction
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Structured Output Example ===")
	fmt.Println("This demonstrates getting JSON-structured responses from Claude.")
	fmt.Println()

	// Show basic structured output
	demonstrateBasicStructuredOutput()

	// Show complex schemas
	demonstrateComplexSchemas()

	// Show Go struct integration
	demonstrateGoStructs()

	// Show parsing patterns
	demonstrateParsingPatterns()

	fmt.Println()
	fmt.Println("=== Structured Output Example Complete ===")
}

// demonstrateBasicStructuredOutput shows basic structured output configuration.
func demonstrateBasicStructuredOutput() {
	fmt.Println("--- Basic Structured Output ---")
	fmt.Println()

	// Simple schema for sentiment analysis
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"sentiment": map[string]any{
				"type":        "string",
				"enum":        []string{"positive", "negative", "neutral"},
				"description": "The overall sentiment of the text",
			},
			"confidence": map[string]any{
				"type":        "number",
				"minimum":     0,
				"maximum":     1,
				"description": "Confidence score from 0 to 1",
			},
			"keywords": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Key terms that influenced the sentiment",
			},
		},
		"required": []string{"sentiment", "confidence"},
	}

	printJSON("Sentiment Analysis Schema", schema)

	fmt.Print(`
  // Configure structured output
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithJSONSchema(schema),
  )

  // Query - response will match the schema
  response, err := client.Query(ctx, "Analyze the sentiment: I love this product!")

  // Parse the structured response
  var result SentimentResult
  json.Unmarshal([]byte(response), &result)
`)
	fmt.Println()
}

// demonstrateComplexSchemas shows more complex schema definitions.
func demonstrateComplexSchemas() {
	fmt.Println("--- Complex Schema Examples ---")
	fmt.Println()

	fmt.Println("1. Code Review Schema:")
	codeReviewSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary": map[string]any{
				"type":        "string",
				"description": "Brief summary of the code review",
			},
			"issues": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"severity": map[string]any{
							"type": "string",
							"enum": []string{"critical", "major", "minor", "suggestion"},
						},
						"line": map[string]any{
							"type":        "integer",
							"description": "Line number where issue occurs",
						},
						"description": map[string]any{
							"type": "string",
						},
						"suggestion": map[string]any{
							"type":        "string",
							"description": "Suggested fix",
						},
					},
					"required": []string{"severity", "description"},
				},
			},
			"score": map[string]any{
				"type":        "integer",
				"minimum":     0,
				"maximum":     100,
				"description": "Overall code quality score",
			},
		},
		"required": []string{"summary", "issues", "score"},
	}
	printJSON("Code Review Schema", codeReviewSchema)

	fmt.Println("2. Entity Extraction Schema:")
	entitySchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"people": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":  map[string]any{"type": "string"},
						"role":  map[string]any{"type": "string"},
						"email": map[string]any{"type": "string", "format": "email"},
					},
					"required": []string{"name"},
				},
			},
			"organizations": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"dates": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"date":    map[string]any{"type": "string", "format": "date"},
						"context": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
	printJSON("Entity Extraction Schema", entitySchema)
	fmt.Println()
}

// demonstrateGoStructs shows integrating with Go structs.
func demonstrateGoStructs() {
	fmt.Println("--- Go Struct Integration ---")
	fmt.Println()

	fmt.Print(`
  // Define your output struct
  type AnalysisResult struct {
      Summary    string   ` + "`json:\"summary\"`" + `
      Categories []string ` + "`json:\"categories\"`" + `
      Confidence float64  ` + "`json:\"confidence\"`" + `
      Metadata   struct {
          ProcessedAt string ` + "`json:\"processed_at\"`" + `
          Model       string ` + "`json:\"model\"`" + `
      } ` + "`json:\"metadata\"`" + `
  }

  // Generate schema from struct (using reflection)
  schema := shared.SchemaFromStruct(AnalysisResult{})

  // Or define schema manually for more control
  schema := map[string]any{
      "type": "object",
      "properties": map[string]any{
          "summary":    map[string]any{"type": "string"},
          "categories": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
          "confidence": map[string]any{"type": "number", "minimum": 0, "maximum": 1},
          "metadata": map[string]any{
              "type": "object",
              "properties": map[string]any{
                  "processed_at": map[string]any{"type": "string", "format": "date-time"},
                  "model":        map[string]any{"type": "string"},
              },
          },
      },
      "required": []string{"summary", "confidence"},
  }
`)
	fmt.Println()
}

// demonstrateParsingPatterns shows patterns for parsing structured output.
func demonstrateParsingPatterns() {
	fmt.Println("--- Parsing Patterns ---")
	fmt.Println()

	// Simulated structured output
	structuredOutput := map[string]any{
		"sentiment":  "positive",
		"confidence": 0.95,
		"keywords":   []string{"love", "great", "excellent"},
	}

	fmt.Println("1. Direct struct parsing:")
	fmt.Print(`
  type SentimentResult struct {
      Sentiment  string   ` + "`json:\"sentiment\"`" + `
      Confidence float64  ` + "`json:\"confidence\"`" + `
      Keywords   []string ` + "`json:\"keywords\"`" + `
  }

  // Get structured output from result message
  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      if result, ok := msg.(*shared.ResultMessage); ok {
          var sentiment SentimentResult
          // StructuredOutput contains the parsed JSON
          data, _ := json.Marshal(result.StructuredOutput)
          json.Unmarshal(data, &sentiment)
          fmt.Printf("Sentiment: %s (%.0f%% confident)\n",
              sentiment.Sentiment, sentiment.Confidence*100)
      }
  }
`)
	fmt.Println()

	printJSON("Example Structured Output", structuredOutput)

	fmt.Println("2. Generic map parsing:")
	fmt.Print(`
  // When schema varies or is dynamic
  if result.StructuredOutput != nil {
      output := result.StructuredOutput.(map[string]any)

      if sentiment, ok := output["sentiment"].(string); ok {
          fmt.Printf("Sentiment: %s\n", sentiment)
      }

      if confidence, ok := output["confidence"].(float64); ok {
          fmt.Printf("Confidence: %.0f%%\n", confidence*100)
      }
  }
`)
	fmt.Println()

	fmt.Println("3. Validation wrapper:")
	fmt.Print(`
  func ParseStructuredOutput[T any](result *shared.ResultMessage) (*T, error) {
      if result.StructuredOutput == nil {
          return nil, fmt.Errorf("no structured output in result")
      }

      data, err := json.Marshal(result.StructuredOutput)
      if err != nil {
          return nil, fmt.Errorf("marshal output: %w", err)
      }

      var output T
      if err := json.Unmarshal(data, &output); err != nil {
          return nil, fmt.Errorf("unmarshal to %T: %w", output, err)
      }

      return &output, nil
  }

  // Usage
  sentiment, err := ParseStructuredOutput[SentimentResult](result)
`)
	fmt.Println()
}

// printJSON prints a labeled JSON object.
func printJSON(label string, v any) {
	data, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Printf("  %s: (error: %v)\n", label, err)
		return
	}
	fmt.Printf("  %s:\n  %s\n\n", label, string(data))
}

// Ensure imports are used
func init() {
	_ = shared.PermissionModeDefault
}

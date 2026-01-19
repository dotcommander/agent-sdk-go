package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBashInput(t *testing.T) {
	input := BashInput{
		Command:     "ls -la",
		Description: "List files",
		Timeout:     5000,
	}
	assert.Equal(t, "ls -la", input.Command)
	assert.Equal(t, "List files", input.Description)
	assert.Equal(t, 5000, input.Timeout)
}

func TestFileEditInput(t *testing.T) {
	input := FileEditInput{
		FilePath:   "/path/to/file.go",
		OldString:  "old",
		NewString:  "new",
		ReplaceAll: true,
	}
	assert.Equal(t, "/path/to/file.go", input.FilePath)
	assert.Equal(t, "old", input.OldString)
	assert.Equal(t, "new", input.NewString)
	assert.True(t, input.ReplaceAll)
}

func TestGrepInput(t *testing.T) {
	input := GrepInput{
		Pattern:    "error",
		Path:       "/src",
		OutputMode: "content",
		HeadLimit:  100,
	}
	assert.Equal(t, "error", input.Pattern)
	assert.Equal(t, "/src", input.Path)
	assert.Equal(t, "content", input.OutputMode)
	assert.Equal(t, 100, input.HeadLimit)
}

func TestTodoWriteInput(t *testing.T) {
	input := TodoWriteInput{
		Todos: []TodoItem{
			{Content: "Task 1", Status: "pending", ActiveForm: "Working on task 1"},
			{Content: "Task 2", Status: "completed", ActiveForm: "Completed task 2"},
		},
	}
	assert.Len(t, input.Todos, 2)
	assert.Equal(t, "pending", input.Todos[0].Status)
}

func TestAskUserQuestionInput(t *testing.T) {
	input := AskUserQuestionInput{
		Questions: []Question{
			{
				Question:    "Choose a framework",
				Header:      "Framework",
				MultiSelect: false,
				Options: []QuestionOption{
					{Label: "React", Description: "UI library"},
					{Label: "Vue", Description: "Progressive framework"},
				},
			},
		},
	}
	assert.Len(t, input.Questions, 1)
	assert.Equal(t, "Framework", input.Questions[0].Header)
}

func TestAgentInputJSON(t *testing.T) {
	input := AgentInput{
		Description:  "Test task",
		Prompt:       "Do something",
		SubagentType: "Explore",
		Model:        "sonnet",
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded AgentInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, input.Description, decoded.Description)
	assert.Equal(t, input.SubagentType, decoded.SubagentType)
}

func TestFileReadInput(t *testing.T) {
	input := FileReadInput{
		FilePath: "/path/to/file.txt",
		Offset:   100,
		Limit:    50,
	}
	assert.Equal(t, "/path/to/file.txt", input.FilePath)
	assert.Equal(t, 100, input.Offset)
	assert.Equal(t, 50, input.Limit)
}

func TestFileWriteInput(t *testing.T) {
	input := FileWriteInput{
		FilePath: "/path/to/output.txt",
		Content:  "Hello, World!",
	}
	assert.Equal(t, "/path/to/output.txt", input.FilePath)
	assert.Equal(t, "Hello, World!", input.Content)
}

func TestGlobInput(t *testing.T) {
	input := GlobInput{
		Pattern: "**/*.go",
		Path:    "/src",
	}
	assert.Equal(t, "**/*.go", input.Pattern)
	assert.Equal(t, "/src", input.Path)
}

func TestWebSearchInput(t *testing.T) {
	input := WebSearchInput{
		Query:          "golang best practices",
		AllowedDomains: []string{"go.dev", "golang.org"},
		BlockedDomains: []string{"spam.com"},
	}
	assert.Equal(t, "golang best practices", input.Query)
	assert.Len(t, input.AllowedDomains, 2)
	assert.Len(t, input.BlockedDomains, 1)
}

func TestWebFetchInput(t *testing.T) {
	input := WebFetchInput{
		URL:    "https://example.com/api",
		Prompt: "Extract the main content",
	}
	assert.Equal(t, "https://example.com/api", input.URL)
	assert.Equal(t, "Extract the main content", input.Prompt)
}

func TestTaskOutputInput(t *testing.T) {
	input := TaskOutputInput{
		TaskID:  "task-123",
		Block:   true,
		Timeout: 30000,
	}
	assert.Equal(t, "task-123", input.TaskID)
	assert.True(t, input.Block)
	assert.Equal(t, 30000, input.Timeout)
}

func TestKillShellInput(t *testing.T) {
	input := KillShellInput{
		ShellID: "shell-456",
	}
	assert.Equal(t, "shell-456", input.ShellID)
}

func TestNotebookEditInput(t *testing.T) {
	input := NotebookEditInput{
		NotebookPath: "/notebooks/analysis.ipynb",
		CellID:       "cell-1",
		CellType:     "code",
		EditMode:     "replace",
		NewSource:    "print('Hello')",
	}
	assert.Equal(t, "/notebooks/analysis.ipynb", input.NotebookPath)
	assert.Equal(t, "cell-1", input.CellID)
	assert.Equal(t, "code", input.CellType)
	assert.Equal(t, "replace", input.EditMode)
	assert.Equal(t, "print('Hello')", input.NewSource)
}

func TestExitPlanModeInput(t *testing.T) {
	input := ExitPlanModeInput{
		AllowedPrompts: []AllowedPrompt{
			{Tool: "Bash", Prompt: "Run tests"},
			{Tool: "Write", Prompt: "Create file"},
		},
	}
	assert.Len(t, input.AllowedPrompts, 2)
	assert.Equal(t, "Bash", input.AllowedPrompts[0].Tool)
}

func TestLSPInput(t *testing.T) {
	input := LSPInput{
		Operation: "goToDefinition",
		FilePath:  "/src/main.go",
		Line:      42,
		Character: 10,
	}
	assert.Equal(t, "goToDefinition", input.Operation)
	assert.Equal(t, "/src/main.go", input.FilePath)
	assert.Equal(t, 42, input.Line)
	assert.Equal(t, 10, input.Character)
}

func TestListMcpResourcesInput(t *testing.T) {
	input := ListMcpResourcesInput{
		ServerName: "my-server",
	}
	assert.Equal(t, "my-server", input.ServerName)
}

func TestReadMcpResourceInput(t *testing.T) {
	input := ReadMcpResourceInput{
		ServerName: "my-server",
		URI:        "file:///path/to/resource",
	}
	assert.Equal(t, "my-server", input.ServerName)
	assert.Equal(t, "file:///path/to/resource", input.URI)
}

func TestConfigInput(t *testing.T) {
	t.Run("get action", func(t *testing.T) {
		input := ConfigInput{
			Action: "get",
			Key:    "model",
		}
		assert.Equal(t, "get", input.Action)
		assert.Equal(t, "model", input.Key)
	})

	t.Run("set action", func(t *testing.T) {
		input := ConfigInput{
			Action: "set",
			Key:    "timeout",
			Value:  30,
			Scope:  "user",
		}
		assert.Equal(t, "set", input.Action)
		assert.Equal(t, "timeout", input.Key)
		assert.Equal(t, 30, input.Value)
		assert.Equal(t, "user", input.Scope)
	})

	t.Run("list action", func(t *testing.T) {
		input := ConfigInput{
			Action: "list",
			Scope:  "project",
		}
		assert.Equal(t, "list", input.Action)
		assert.Equal(t, "project", input.Scope)
	})
}

func TestQuestionOption(t *testing.T) {
	opt := QuestionOption{
		Label:       "Option A",
		Description: "First option",
	}
	assert.Equal(t, "Option A", opt.Label)
	assert.Equal(t, "First option", opt.Description)
}

func TestQuestionMetadata(t *testing.T) {
	meta := QuestionMetadata{
		Source: "user_input",
	}
	assert.Equal(t, "user_input", meta.Source)
}

func TestAllowedPrompt(t *testing.T) {
	prompt := AllowedPrompt{
		Tool:   "Bash",
		Prompt: "Run npm install",
	}
	assert.Equal(t, "Bash", prompt.Tool)
	assert.Equal(t, "Run npm install", prompt.Prompt)
}

func TestBashInputJSON(t *testing.T) {
	input := BashInput{
		Command:                   "ls -la",
		Description:               "List files",
		Timeout:                   5000,
		RunInBackground:           true,
		DangerouslyDisableSandbox: false,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded BashInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, input.Command, decoded.Command)
	assert.Equal(t, input.Description, decoded.Description)
	assert.Equal(t, input.Timeout, decoded.Timeout)
	assert.True(t, decoded.RunInBackground)
	assert.False(t, decoded.DangerouslyDisableSandbox)
}

func TestTodoItemStatuses(t *testing.T) {
	statuses := []string{"pending", "in_progress", "completed"}
	for _, status := range statuses {
		item := TodoItem{
			Content:    "Test task",
			Status:     status,
			ActiveForm: "Testing task",
		}
		assert.Equal(t, "Test task", item.Content)
		assert.Equal(t, status, item.Status)
		assert.Equal(t, "Testing task", item.ActiveForm)
	}
}

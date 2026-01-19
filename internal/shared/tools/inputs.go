package tools

// AgentInput represents input for the Task (Agent) tool.
type AgentInput struct {
	Description     string `json:"description"`
	Prompt          string `json:"prompt"`
	SubagentType    string `json:"subagent_type"`
	Model           string `json:"model,omitempty"`
	MaxTurns        int    `json:"max_turns,omitempty"`
	Resume          string `json:"resume,omitempty"`
	RunInBackground bool   `json:"run_in_background,omitempty"`
}

// BashInput represents input for the Bash tool.
type BashInput struct {
	Command                   string `json:"command"`
	Description               string `json:"description,omitempty"`
	Timeout                   int    `json:"timeout,omitempty"`
	RunInBackground           bool   `json:"run_in_background,omitempty"`
	DangerouslyDisableSandbox bool   `json:"dangerouslyDisableSandbox,omitempty"`
}

// TaskOutputInput represents input for the TaskOutput tool.
type TaskOutputInput struct {
	TaskID  string `json:"task_id"`
	Block   bool   `json:"block"`
	Timeout int    `json:"timeout,omitempty"`
}

// FileEditInput represents input for the Edit tool.
type FileEditInput struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// FileReadInput represents input for the Read tool.
type FileReadInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// FileWriteInput represents input for the Write tool.
type FileWriteInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// GlobInput represents input for the Glob tool.
type GlobInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// GrepInput represents input for the Grep tool.
type GrepInput struct {
	Pattern         string `json:"pattern"`
	Path            string `json:"path,omitempty"`
	Glob            string `json:"glob,omitempty"`
	Type            string `json:"type,omitempty"`
	OutputMode      string `json:"output_mode,omitempty"`
	HeadLimit       int    `json:"head_limit,omitempty"`
	Offset          int    `json:"offset,omitempty"`
	Multiline       bool   `json:"multiline,omitempty"`
	CaseInsensitive bool   `json:"-i,omitempty"`
	ShowLineNumbers bool   `json:"-n,omitempty"`
	ContextBefore   int    `json:"-B,omitempty"`
	ContextAfter    int    `json:"-A,omitempty"`
	ContextAround   int    `json:"-C,omitempty"`
}

// KillShellInput represents input for the KillShell tool.
type KillShellInput struct {
	ShellID string `json:"shell_id"`
}

// NotebookEditInput represents input for the NotebookEdit tool.
type NotebookEditInput struct {
	NotebookPath string `json:"notebook_path"`
	CellID       string `json:"cell_id,omitempty"`
	CellType     string `json:"cell_type,omitempty"`
	EditMode     string `json:"edit_mode,omitempty"`
	NewSource    string `json:"new_source"`
}

// WebFetchInput represents input for the WebFetch tool.
type WebFetchInput struct {
	URL    string `json:"url"`
	Prompt string `json:"prompt"`
}

// WebSearchInput represents input for the WebSearch tool.
type WebSearchInput struct {
	Query          string   `json:"query"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
}

// TodoWriteInput represents input for the TodoWrite tool.
type TodoWriteInput struct {
	Todos []TodoItem `json:"todos"`
}

// TodoItem represents a single todo item.
type TodoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"` // "pending" | "in_progress" | "completed"
	ActiveForm string `json:"activeForm"`
}

// AskUserQuestionInput represents input for the AskUserQuestion tool.
type AskUserQuestionInput struct {
	Questions []Question        `json:"questions"`
	Answers   map[string]string `json:"answers,omitempty"`
	Metadata  *QuestionMetadata `json:"metadata,omitempty"`
}

// Question represents a question in AskUserQuestion.
type Question struct {
	Question    string           `json:"question"`
	Header      string           `json:"header"`
	Options     []QuestionOption `json:"options"`
	MultiSelect bool             `json:"multiSelect"`
}

// QuestionOption represents an option in a question.
type QuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

// QuestionMetadata contains optional metadata for questions.
type QuestionMetadata struct {
	Source string `json:"source,omitempty"`
}

// ExitPlanModeInput represents input for the ExitPlanMode tool.
type ExitPlanModeInput struct {
	AllowedPrompts []AllowedPrompt `json:"allowedPrompts,omitempty"`
}

// AllowedPrompt represents a permission request for plan mode.
type AllowedPrompt struct {
	Tool   string `json:"tool"`
	Prompt string `json:"prompt"`
}

// LSPInput represents input for the LSP tool.
type LSPInput struct {
	Operation string `json:"operation"`
	FilePath  string `json:"filePath"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

// ListMcpResourcesInput represents input for the ListMcpResources tool.
type ListMcpResourcesInput struct {
	ServerName string `json:"server_name,omitempty"`
}

// ReadMcpResourceInput represents input for the ReadMcpResource tool.
type ReadMcpResourceInput struct {
	ServerName string `json:"server_name"`
	URI        string `json:"uri"`
}

// ConfigInput represents input for the Config tool.
type ConfigInput struct {
	Action string `json:"action"` // "get" | "set" | "list"
	Key    string `json:"key,omitempty"`
	Value  any    `json:"value,omitempty"`
	Scope  string `json:"scope,omitempty"` // "local" | "user" | "project"
}

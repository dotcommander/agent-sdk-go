package shared

// PermissionMode controls how tool executions are handled.
type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
	PermissionModePlan              PermissionMode = "plan"
	PermissionModeDelegate          PermissionMode = "delegate"
	PermissionModeDontAsk           PermissionMode = "dontAsk"
)

// PermissionBehavior determines how a permission request is handled.
type PermissionBehavior string

const (
	PermissionBehaviorAllow PermissionBehavior = "allow"
	PermissionBehaviorDeny  PermissionBehavior = "deny"
	PermissionBehaviorAsk   PermissionBehavior = "ask"
)

// PermissionUpdateDestination specifies where permission updates are stored.
type PermissionUpdateDestination string

const (
	PermissionDestUserSettings    PermissionUpdateDestination = "userSettings"
	PermissionDestProjectSettings PermissionUpdateDestination = "projectSettings"
	PermissionDestLocalSettings   PermissionUpdateDestination = "localSettings"
	PermissionDestSession         PermissionUpdateDestination = "session"
	PermissionDestCLIArg          PermissionUpdateDestination = "cliArg"
)

// PermissionRuleValue represents a permission rule.
type PermissionRuleValue struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// PermissionUpdate represents an update to permission configuration.
// This is a discriminated union based on the Type field.
type PermissionUpdate struct {
	Type        string                      `json:"type"` // "addRules", "replaceRules", "removeRules", "setMode", "addDirectories", "removeDirectories"
	Rules       []PermissionRuleValue       `json:"rules,omitempty"`
	Behavior    PermissionBehavior          `json:"behavior,omitempty"`
	Destination PermissionUpdateDestination `json:"destination,omitempty"`
	Mode        PermissionMode              `json:"mode,omitempty"`
	Directories []string                    `json:"directories,omitempty"`
}

// PermissionResult represents the result of a permission check.
// This is a discriminated union based on the Behavior field.
type PermissionResult struct {
	Behavior           PermissionBehavior `json:"behavior"` // "allow" or "deny"
	UpdatedInput       map[string]any     `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updatedPermissions,omitempty"`
	ToolUseID          string             `json:"toolUseID,omitempty"`
	Message            string             `json:"message,omitempty"`   // for deny
	Interrupt          bool               `json:"interrupt,omitempty"` // for deny
}

// CanUseToolOptions contains options passed to the CanUseTool callback.
type CanUseToolOptions struct {
	Suggestions    []PermissionUpdate `json:"suggestions,omitempty"`
	BlockedPath    string             `json:"blockedPath,omitempty"`
	DecisionReason string             `json:"decisionReason,omitempty"`
	ToolUseID      string             `json:"toolUseID"`
	AgentID        string             `json:"agentID,omitempty"`
}

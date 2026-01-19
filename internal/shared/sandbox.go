package shared

// SandboxNetworkConfig configures network access in sandbox.
type SandboxNetworkConfig struct {
	AllowedDomains      []string `json:"allowedDomains,omitempty"`
	AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
	HttpProxyPort       int      `json:"httpProxyPort,omitempty"`
	SocksProxyPort      int      `json:"socksProxyPort,omitempty"`
}

// RipgrepConfig configures ripgrep in sandbox.
type RipgrepConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// SandboxConfig provides comprehensive sandbox configuration.
// This extends the basic SandboxSettings with additional isolation options.
type SandboxConfig struct {
	Enabled                   bool                `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  bool                `json:"autoAllowBashIfSandboxed,omitempty"`
	AllowUnsandboxedCommands  bool                `json:"allowUnsandboxedCommands,omitempty"`
	Network                   *SandboxNetworkConfig `json:"network,omitempty"`
	IgnoreViolations          map[string][]string `json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox bool                `json:"enableWeakerNestedSandbox,omitempty"`
	ExcludedCommands          []string            `json:"excludedCommands,omitempty"`
	Ripgrep                   *RipgrepConfig      `json:"ripgrep,omitempty"`
}

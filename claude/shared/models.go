package shared

// Model short name constants.
// These match the TypeScript SDK's AgentDefinition which accepts 'sonnet' | 'opus' | 'haiku' | 'inherit'.
const (
	ModelShortNameOpus   = "opus"
	ModelShortNameSonnet = "sonnet"
	ModelShortNameHaiku  = "haiku"
)

// modelShortNames maps short model names to their full model IDs.
var modelShortNames = map[string]string{
	ModelShortNameOpus:   "claude-opus-4-5-20251101",
	ModelShortNameSonnet: "claude-sonnet-4-5-20250929",
	ModelShortNameHaiku:  "claude-3-5-haiku-20241022",
}

// ResolveModelName resolves a short model name to its full model ID.
// If the input is already a full model ID (or unknown), it is returned unchanged.
func ResolveModelName(name string) string {
	if fullName, ok := modelShortNames[name]; ok {
		return fullName
	}
	return name
}

// IsValidModelShortName reports whether name is a recognized short model name.
func IsValidModelShortName(name string) bool {
	_, ok := modelShortNames[name]
	return ok
}

// GetModelShortNames returns all recognized short model names.
func GetModelShortNames() []string {
	names := make([]string, 0, len(modelShortNames))
	for name := range modelShortNames {
		names = append(names, name)
	}
	return names
}

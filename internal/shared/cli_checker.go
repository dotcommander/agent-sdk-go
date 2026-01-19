package shared

// CLIChecker defines the interface for checking CLI availability.
// This enables dependency injection for testability - tests can inject
// a mock that always returns true/false without requiring the actual CLI.
type CLIChecker interface {
	IsCLIAvailable() bool
}

// CLICheckerFunc is a function type that implements CLIChecker.
// This allows for easy inline checker creation in tests.
type CLICheckerFunc func() bool

// IsCLIAvailable implements CLIChecker.
func (f CLICheckerFunc) IsCLIAvailable() bool {
	return f()
}

// AlwaysAvailableCLIChecker is a CLIChecker that always returns true.
// Useful for testing without the actual CLI.
type AlwaysAvailableCLIChecker struct{}

// IsCLIAvailable implements CLIChecker.
func (AlwaysAvailableCLIChecker) IsCLIAvailable() bool {
	return true
}

// NeverAvailableCLIChecker is a CLIChecker that always returns false.
// Useful for testing CLI unavailability scenarios.
type NeverAvailableCLIChecker struct{}

// IsCLIAvailable implements CLIChecker.
func (NeverAvailableCLIChecker) IsCLIAvailable() bool {
	return false
}

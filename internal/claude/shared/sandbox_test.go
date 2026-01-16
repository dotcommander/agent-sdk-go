package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxConfig(t *testing.T) {
	config := SandboxConfig{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
		Network: &SandboxNetworkConfig{
			AllowedDomains:    []string{"example.com"},
			AllowLocalBinding: true,
		},
	}
	assert.True(t, config.Enabled)
	assert.True(t, config.AutoAllowBashIfSandboxed)
	assert.NotNil(t, config.Network)
	assert.Equal(t, []string{"example.com"}, config.Network.AllowedDomains)
}

func TestSandboxSettings(t *testing.T) {
	settings := SandboxSettings{
		Enabled:    true,
		Type:       "docker",
		Image:      "ubuntu:latest",
		WorkingDir: "/app",
		Options:    map[string]string{"memory": "512m"},
	}
	assert.True(t, settings.Enabled)
	assert.Equal(t, "docker", settings.Type)
	assert.Equal(t, "ubuntu:latest", settings.Image)
	assert.Equal(t, "/app", settings.WorkingDir)
	assert.Equal(t, "512m", settings.Options["memory"])
}

func TestRipgrepConfig(t *testing.T) {
	config := RipgrepConfig{
		Command: "rg",
		Args:    []string{"--ignore-case"},
	}
	assert.Equal(t, "rg", config.Command)
	assert.Equal(t, []string{"--ignore-case"}, config.Args)
}

func TestSandboxNetworkConfig(t *testing.T) {
	config := SandboxNetworkConfig{
		AllowedDomains:      []string{"api.example.com", "cdn.example.com"},
		AllowUnixSockets:    []string{"/var/run/docker.sock"},
		AllowAllUnixSockets: false,
		AllowLocalBinding:   true,
		HttpProxyPort:       8080,
		SocksProxyPort:      1080,
	}
	assert.Equal(t, []string{"api.example.com", "cdn.example.com"}, config.AllowedDomains)
	assert.Equal(t, []string{"/var/run/docker.sock"}, config.AllowUnixSockets)
	assert.False(t, config.AllowAllUnixSockets)
	assert.True(t, config.AllowLocalBinding)
	assert.Equal(t, 8080, config.HttpProxyPort)
	assert.Equal(t, 1080, config.SocksProxyPort)
}

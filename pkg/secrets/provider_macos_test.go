//go:build darwin

/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package secrets

import (
	"os/exec"
	"strings"
	"testing"
)

// isMacOS checks if the current platform is macOS
func isMacOS() bool {
	cmd := exec.Command("uname", "-s")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "Darwin"
}

// Platform-specific keychain tests can go here
func TestRealSecrets_buildServiceName(t *testing.T) {
	r := NewRealSecrets("testapp")
	got := r.buildServiceName("auth")
	want := "cli.testapp.auth"
	if got != want {
		t.Errorf("buildServiceName() = %q, want %q", got, want)
	}
}

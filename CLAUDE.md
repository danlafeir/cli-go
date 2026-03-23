# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

**cli-go** is a shared utilities library for building CLI tools in Go. It auto-discovers executables matching a configurable prefix (e.g., `myapp-*`) from PATH and registers them as dynamic subcommands (plugin architecture). It also provides built-in packages for configuration management, secrets management, and self-update.

## Commands

```bash
# Build for current platform
make build          # outputs to bin/cli-go

# Cross-compile for all platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
make build-all

# Run all tests
make test           # runs go test ./...

# Run a single test
go test ./pkg/config/... -run TestFunctionName

# Clean build artifacts
make clean

# Run without building
go run main.go [command]
```

## Architecture

### Plugin System (`pkg/plugin/`)
Scans PATH for executables matching a caller-supplied prefix (e.g., `myapp-`). Each match is registered as a Cobra subcommand that proxies args and I/O to the subprocess. This is the primary extension mechanism. Call `RegisterPlugins(rootCmd, "myapp-")` to register plugins.

### Configuration (`pkg/config/`)
YAML-based config stored in a caller-supplied directory. Uses Viper under the hood with namespaced key-value storage (`cmd.key.subkey`). Schema validation supports wildcard patterns (`*` matches a single segment). Key entry points: `InitConfig(path)`, `GetConfigValue()`, `SetConfigValue()`. `InitConfig` requires a non-empty path — callers own their config dir location.

### Secrets (`pkg/secrets/`)
Cross-platform secrets management using OS-native keychains. Platform-specific implementations are selected at compile time via build tags:
- `provider_macos.go` — `//go:build darwin` — macOS Keychain via `go-keychain`
- `provider_linux.go` — `//go:build linux` — Linux Secret Service via dbus

Uses a `SecretsProvider` interface. Instantiate with `NewRealSecrets(appName)`. Call `SetDefaultProvider(appName)` to initialize package-level functions. Keys follow the convention `cli.<appName>.<namespace>.<name>`.

### Self-Update (`pkg/update/`)
Queries GitHub releases API for the latest binary matching the current OS/ARCH. Downloads and atomically replaces the running binary. Falls back to `sudo` if write permission is denied. Callers construct their own `Config{AppName: ..., Repo: ...}` and call `RunUpdateWithConfig`.

### Test Utilities (`testutil/mocks/`)
`mock_secrets.go` provides an in-memory `SecretsProvider` implementation for unit tests. Instantiate with `NewMockSecrets(appName)`. Use this instead of real keychain calls in tests.

### Entry Point
`main.go` initializes the Cobra root command, registers built-in subcommands from `cmd/`, and calls the plugin discovery logic before executing.

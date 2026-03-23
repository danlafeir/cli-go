# cli-go

A Go library of shared utilities for building CLI tools. Import the packages you need:

- **`pkg/plugin`** — auto-discover and register executables from PATH as Cobra subcommands
- **`pkg/config`** — YAML-based config with namespaced key-value storage (backed by Viper)
- **`pkg/secrets`** — OS-native keychain secrets management (macOS Keychain, Linux Secret Service)
- **`pkg/update`** — self-update by querying GitHub releases for the latest binary

## Usage

```go
import (
    "github.com/danlafeir/cli-go/pkg/plugin"
    "github.com/danlafeir/cli-go/pkg/config"
    "github.com/danlafeir/cli-go/pkg/secrets"
    "github.com/danlafeir/cli-go/pkg/update"
)

// Register plugins matching "myapp-*" from PATH
plugin.RegisterPlugins(rootCmd, "myapp-")

// Initialize config from a directory
config.InitConfig(filepath.Join(os.UserHomeDir(), ".myapp"))

// Read/write secrets scoped to your app
secrets.SetDefaultProvider("myapp")
secrets.Write("auth", "token", "my-secret-value")

// Check for and apply updates
update.RunUpdateWithConfig(update.Config{AppName: "myapp", Repo: "org/myapp"}, currentHash, cmd)
```

## Development

```bash
# Run tests
make test

# Build
make build

# Cross-compile (e.g. Linux amd64)
GOOS=linux GOARCH=amd64 make build
```

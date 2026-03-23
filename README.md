# cli-go

A shared utilities library for building CLI tools in Go. Provides plugin discovery, configuration management, secrets management, and self-update capabilities.

## Getting Started

### Run Tests

To run all tests:

```
go test ./...
```

### Run Locally

To run the CLI locally (from the project root):

```
go run main.go [command]
```

For example, to see available commands:

```
go run main.go --help
```

### Build

To build the project for your current OS/architecture:

```
make build
```

This will produce a binary named `cli-go` in the `bin/` directory for your current system.

To cross-compile for another system (e.g., Linux amd64):

```
GOOS=linux GOARCH=amd64 make build
```

This will also produce a binary named `cli-go` in the `bin/` directory, but for the specified target system.

## Quick Install (Online Script)

You can install the latest `cli-go` binary automatically with a one-liner (Linux/macOS):

```sh
curl -sSL https://raw.githubusercontent.com/danlafeir/cli-go/main/scripts/install.sh | sh
```

This script will detect your OS and architecture, download the correct binary from the main branch, and install it to `~/.local/bin` (you may be prompted for your password).

**Security tip:** Always review install scripts before piping to `sh`.
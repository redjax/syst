# syst CLI Tool - AI Agent Instructions

## Architecture Overview

**syst** is a cross-platform Go CLI utility with a modular command/service architecture. The core pattern follows:
- `cmd/` - Root Cobra command definition and entrypoint
- `internal/commands/` - Command implementations organized by domain 
- `internal/services/` - Business logic services supporting commands
- `internal/utils/` - Shared utilities and helpers

### Command Registration Pattern
Commands follow a factory pattern where each command directory provides a `New*Command()` function that returns a configured `*cobra.Command`. These are registered in `cmd/root.go`:

```go
rootCmd.AddCommand(showCommand.NewShowCmd())
rootCmd.AddCommand(_git.NewGitCommand())
// etc.
```

### Service-Command Separation
Commands act as thin CLI adapters that delegate to services:
- **Commands** handle CLI concerns: flags, help text, argument parsing
- **Services** contain business logic and can be unit tested independently
- Example: `pingCommand.NewPingCommand()` → `pingService.RunPing()`

## Key Conventions

### Cross-Platform Service Implementation
Services handle platform differences using build tags:
```go
//go:build !windows
// +build !windows
func runICMPPing() { /* Unix implementation */ }

//go:build windows 
// +build windows  
func runICMPPing() { /* Windows implementation */ }
```

### Configuration Management
Uses [koanf](https://github.com/knadh/koanf) for hierarchical config with precedence:
1. Command-line flags (highest)
2. Environment variables (`SYST_*` prefix)
3. Config file (JSON/YAML/TOML)
4. Defaults (lowest)

Environment variables transform: `SYST_FOO_BAR` → `foo.bar` in config.

### Version/Build Metadata Injection
Build scripts inject version info via ldflags:
```bash
-X 'github.com/redjax/syst/internal/version.Version=${GitVersion}'
-X 'github.com/redjax/syst/internal/version.Commit=${GitCommit}'
-X 'github.com/redjax/syst/internal/version.Date=${BuildDate}'
```

## Development Workflows

### Building
Use provided build scripts instead of raw `go build`:
- **Linux/Mac**: `./scripts/build/build.sh --bin-name syst --build-os linux`
- **Windows**: `.\scripts\build\build.ps1 -BinName syst -BuildOS windows`
- **Quick dev build**: `go build -o build/syst ./cmd/entrypoint`

Default entrypoint: `./cmd/entrypoint/main.go`

### Adding New Commands
1. Create directory in `internal/commands/` 
2. Implement `New*Command() *cobra.Command` factory
3. Create corresponding service in `internal/services/` if needed
4. Register in `cmd/root.go` init function
5. Add README.md explaining usage

### Service Structure Patterns
- **Single-file services**: Simple utilities like `whichCommand` 
- **Multi-file services**: Complex domains like `gitService/` and `pingService/`
- **Platform services**: Use build tags for OS-specific implementations
- **State management**: Pass context and options structs rather than globals

## Key Dependencies & Integrations

### UI/Display Libraries
- **Cobra**: CLI framework - all commands must return `*cobra.Command`
- **go-pretty/table**: Formatted table output (see `pingService.PrettyPrintPingSummaryTable`)
- **Bubble Tea + Lipgloss**: TUI applications (see `gitService/activity/`)

### Platform Integrations
- **gopsutil**: System information gathering
- **go-git**: Git repository operations  
- **pro-bing**: Cross-platform ping functionality
- Platform-specific syscalls in `platformService/`

### Self-Management Features
The CLI includes self-upgrade capabilities via `version.NewSelfCommand()` - downloads and replaces binary from GitHub releases.

## Testing & Debugging
- Global debug flag: `syst --debug` or `syst -D` 
- Most services accept context for cancellation
- Services often include logging options (see `pingService.Options.LogToFile`)
- Use `cmd.Help()` as fallback when no args provided

## Common Patterns to Follow
- **Error handling**: Wrap with context using `fmt.Errorf("operation failed: %w", err)`
- **Contexts**: Pass `context.Context` for cancellable operations
- **Options structs**: Use for complex service configurations 
- **Platform detection**: Use `runtime.GOOS` and `platformService` constants
- **Output**: Support both simple text and formatted table output modes
